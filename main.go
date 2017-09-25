package main

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"tisupvisor/metrics"
)

func main() {
	fmt.Println("start...")
	ctx, cancel := context.WithCancel(context.Background())

	sc := make(chan os.Signal, 1)
	signal.Notify(sc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		sig := <-sc
		log.Infof("Got signal [%d] to exit.", sig)
		cancel()
		os.Exit(0)
	}()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		metrics.InstanceMetrics(ctx, "/tmp", 8081, 30)
	}()
	wg.Wait()
	return

}
