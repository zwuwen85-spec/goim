package job

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	pb "github.com/Terry-Mao/goim/api/logic"
	"github.com/Terry-Mao/goim/internal/job/conf"
	"github.com/bilibili/discovery/naming"
	"github.com/golang/protobuf/proto"

	cluster "github.com/bsm/sarama-cluster"
	log "github.com/golang/glog"
)

// Job is push job.
type Job struct {
	c            *conf.Config
	consumer     *cluster.Consumer
	cometServers map[string]*Comet

	rooms      map[string]*Room
	roomsMutex sync.RWMutex
}

// New new a push job.
func New(c *conf.Config) *Job {
	fmt.Fprintf(os.Stderr, "=== JOB: New() START ===\n")
	j := &Job{
		c:        c,
		consumer: newKafkaSub(c.Kafka),
		rooms:    make(map[string]*Room),
	}
	fmt.Fprintf(os.Stderr, "=== JOB: New() about to call watchComet ===\n")
	j.watchComet(c.Discovery)
	fmt.Fprintf(os.Stderr, "=== JOB: New() watchComet returned, returning Job ===\n")
	return j
}

func newKafkaSub(c *conf.Kafka) *cluster.Consumer {
	fmt.Fprintf(os.Stderr, "=== JOB: Creating Kafka consumer: brokers=%v topic=%s group=%s ===\n", c.Brokers, c.Topic, c.Group)
	config := cluster.NewConfig()
	config.Consumer.Return.Errors = true
	config.Group.Return.Notifications = true
	consumer, err := cluster.NewConsumer(c.Brokers, c.Group, []string{c.Topic}, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "=== JOB: Kafka consumer creation FAILED: %v ===\n", err)
		panic(err)
	}
	fmt.Fprintf(os.Stderr, "=== JOB: Kafka consumer created successfully ===\n")
	return consumer
}

// Close close resounces.
func (j *Job) Close() error {
	if j.consumer != nil {
		return j.consumer.Close()
	}
	return nil
}

// Consume messages, watch signals
func (j *Job) Consume() {
	fmt.Fprintf(os.Stderr, "=== JOB: Consume() starting ===\n")
	msgCount := 0
	for {
		select {
		case err := <-j.consumer.Errors():
			log.Errorf("consumer error(%v)", err)
			fmt.Fprintf(os.Stderr, "=== JOB: Consumer error: %v ===\n", err)
		case n := <-j.consumer.Notifications():
			log.Infof("consumer rebalanced(%v)", n)
			fmt.Fprintf(os.Stderr, "=== JOB: Consumer rebalanced: %v ===\n", n)
		case msg, ok := <-j.consumer.Messages():
			msgCount++
			fmt.Fprintf(os.Stderr, "=== JOB: Message #%d received: topic=%s partition=%d offset=%d key=%s ===\n", msgCount, msg.Topic, msg.Partition, msg.Offset, msg.Key)
			if !ok {
				fmt.Fprintf(os.Stderr, "=== JOB: Consumer channel closed, exiting ===\n")
				return
			}
			j.consumer.MarkOffset(msg, "")
			// process push message
			pushMsg := new(pb.PushMsg)
			if err := proto.Unmarshal(msg.Value, pushMsg); err != nil {
				log.Errorf("proto.Unmarshal(%v) error(%v)", msg, err)
				fmt.Fprintf(os.Stderr, "=== JOB: Unmarshal error: %v ===\n", err)
				continue
			}
			fmt.Fprintf(os.Stderr, "=== JOB: PushMsg parsed: Type=%d Server=%s Keys=%v ===\n", pushMsg.Type, pushMsg.Server, pushMsg.Keys)
			if err := j.push(context.Background(), pushMsg); err != nil {
				log.Errorf("j.push(%v) error(%v)", pushMsg, err)
				fmt.Fprintf(os.Stderr, "=== JOB: Push error: %v ===\n", err)
			} else {
				fmt.Fprintf(os.Stderr, "=== JOB: Push SUCCESS ===\n")
			}
			log.Infof("consume: %s/%d/%d\t%s\t%+v", msg.Topic, msg.Partition, msg.Offset, msg.Key, pushMsg)
		}
	}
}

func (j *Job) watchComet(c *naming.Config) {
	fmt.Fprintf(os.Stderr, "=== JOB: watchComet() START ===\n")
	dis := naming.New(c)
	fmt.Fprintf(os.Stderr, "=== JOB: watchComet() discovery created ===\n")
	resolver := dis.Build("goim.comet")
	fmt.Fprintf(os.Stderr, "=== JOB: watchComet() resolver built for goim.comet ===\n")
	event := resolver.Watch()
	fmt.Fprintf(os.Stderr, "=== JOB: watchComet() about to enter initial select ===\n")
	select {
	case _, ok := <-event:
		if !ok {
			panic("watchComet init failed")
		}
		if ins, ok := resolver.Fetch(); ok {
			if err := j.newAddress(ins.Instances); err != nil {
				panic(err)
			}
			log.Infof("watchComet init newAddress:%+v", ins)
		}
	case <-time.After(10 * time.Second):
		log.Error("watchComet init instances timeout")
	}
	fmt.Fprintf(os.Stderr, "=== JOB: watchComet() initial select completed, starting goroutine ===\n")
	go func() {
		for {
			if _, ok := <-event; !ok {
				log.Info("watchComet exit")
				return
			}
			ins, ok := resolver.Fetch()
			if ok {
				if err := j.newAddress(ins.Instances); err != nil {
					log.Errorf("watchComet newAddress(%+v) error(%+v)", ins, err)
					continue
				}
				log.Infof("watchComet change newAddress:%+v", ins)
			}
		}
	}()
}

func (j *Job) newAddress(insMap map[string][]*naming.Instance) error {
	ins := insMap[j.c.Env.Zone]
	if len(ins) == 0 {
		return fmt.Errorf("watchComet instance is empty")
	}
	comets := map[string]*Comet{}
	for _, in := range ins {
		if old, ok := j.cometServers[in.Hostname]; ok {
			comets[in.Hostname] = old
			continue
		}
		c, err := NewComet(in, j.c.Comet)
		if err != nil {
			log.Errorf("watchComet NewComet(%+v) error(%v)", in, err)
			return err
		}
		comets[in.Hostname] = c
		log.Infof("watchComet AddComet grpc:%+v", in)
	}
	for key, old := range j.cometServers {
		if _, ok := comets[key]; !ok {
			old.cancel()
			log.Infof("watchComet DelComet:%s", key)
		}
	}
	j.cometServers = comets
	return nil
}
