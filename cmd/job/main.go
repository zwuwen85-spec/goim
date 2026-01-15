package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bilibili/discovery/naming"
	"github.com/Terry-Mao/goim/internal/job"
	"github.com/Terry-Mao/goim/internal/job/conf"

	resolver "github.com/bilibili/discovery/naming/grpc"
	log "github.com/golang/glog"
)

var (
	ver = "2.0.0"
)

func main() {
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	log.Infof("goim-job [version: %s env: %+v] start", ver, conf.Conf.Env)
	fmt.Fprintf(os.Stderr, "=== JOB MAIN: About to create discovery and register resolver ===\n")
	// grpc register naming
	dis := naming.New(conf.Conf.Discovery)
	resolver.Register(dis)
	fmt.Fprintf(os.Stderr, "=== JOB MAIN: About to call job.New() ===\n")
	// job
	j := job.New(conf.Conf)
	fmt.Fprintf(os.Stderr, "=== JOB MAIN: job.New() returned, about to start Consume goroutine ===\n")
	go j.Consume()
	fmt.Fprintf(os.Stderr, "=== JOB MAIN: Consume goroutine started, entering signal loop ===\n")
	// signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Infof("goim-job get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			j.Close()
			log.Infof("goim-job [version: %s] exit", ver)
			log.Flush()
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
