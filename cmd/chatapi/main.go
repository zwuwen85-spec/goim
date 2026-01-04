package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Terry-Mao/goim/internal/chatapi"
	"github.com/Terry-Mao/goim/internal/chatapi/conf"
	log "github.com/golang/glog"
)

var (
	// Build info
	version   string
	buildTime string
)

func main() {
	flag.Parse()
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s config_file.toml\n", os.Args[0])
		os.Exit(1)
	}

	// Load config
	cfgPath := os.Args[1]
	if err := conf.Init(cfgPath); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Infof("ChatAPI starting [version: %s, build: %s]", version, buildTime)

	// Create server
	srv, err := chatapi.NewServer(conf.Conf)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Start server
	if err := srv.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
