package comet

import (
	"fmt"
	"os"
	"sync"

	"github.com/Terry-Mao/goim/api/protocol"
	"github.com/Terry-Mao/goim/internal/comet/errors"
	"github.com/Terry-Mao/goim/pkg/bufio"
)

// Channel used by message pusher send msg to write goroutine.
type Channel struct {
	Room     *Room
	CliProto Ring
	signal   chan *protocol.Proto
	Writer   bufio.Writer
	Reader   bufio.Reader
	Next     *Channel
	Prev     *Channel

	Mid      int64
	Key      string
	IP       string
	watchOps map[int32]struct{}
	mutex    sync.RWMutex
}

// NewChannel new a channel.
func NewChannel(cli, svr int) *Channel {
	c := new(Channel)
	c.CliProto.Init(cli)
	c.signal = make(chan *protocol.Proto, svr)
	c.watchOps = make(map[int32]struct{})
	return c
}

// Watch watch a operation.
func (c *Channel) Watch(accepts ...int32) {
	fmt.Fprintf(os.Stderr, "=== Channel Watch: accepts=%v ===\n", accepts)
	c.mutex.Lock()
	for _, op := range accepts {
		c.watchOps[op] = struct{}{}
	}
	c.mutex.Unlock()
	fmt.Fprintf(os.Stderr, "=== Channel Watch: watchOps now has %d items ===\n", len(c.watchOps))
}

// UnWatch unwatch an operation
func (c *Channel) UnWatch(accepts ...int32) {
	c.mutex.Lock()
	for _, op := range accepts {
		delete(c.watchOps, op)
	}
	c.mutex.Unlock()
}

// NeedPush verify if in watch.
func (c *Channel) NeedPush(op int32) bool {
	c.mutex.RLock()
	if _, ok := c.watchOps[op]; ok {
		c.mutex.RUnlock()
		return true
	}
	c.mutex.RUnlock()
	return false
}

// Push server push message.
func (c *Channel) Push(p *protocol.Proto) (err error) {
	fmt.Fprintf(os.Stderr, "=== Channel Push: key=%s protoOp=%d bodyLen=%d ===\n", c.Key, p.Op, len(p.Body))
	select {
	case c.signal <- p:
		fmt.Fprintf(os.Stderr, "=== Channel Push: SUCCESS sent to signal ===\n")
	default:
		fmt.Fprintf(os.Stderr, "=== Channel Push: FAILED signal full ===\n")
		err = errors.ErrSignalFullMsgDropped
	}
	return
}

// Ready check the channel ready or close?
func (c *Channel) Ready() *protocol.Proto {
	return <-c.signal
}

// Signal send signal to the channel, protocol ready.
func (c *Channel) Signal() {
	c.signal <- protocol.ProtoReady
}

// Close close the channel.
func (c *Channel) Close() {
	c.signal <- protocol.ProtoFinish
}
