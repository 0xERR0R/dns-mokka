package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/0xERR0R/dns-mokka/config"
	"github.com/0xERR0R/dns-mokka/server"
	log "github.com/sirupsen/logrus"
)

func main() {
	cfg, err := config.ReadConfig()

	if err != nil {
		log.Fatal("can't read configuration: ", err)
	}

	srv, err := server.NewServer(cfg)

	if err != nil {
		log.Fatal("can't create DNS server: ", err)
	}

	signals := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	srv.Start()

	go func() {
		<-signals
		log.Infof("Terminating...")
		srv.Stop()
		done <- true
	}()

	<-done
}
