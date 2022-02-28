package main

import (
	"fmt"
	"github.com/foximilUno/metrics/internal/collector"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type config struct {
	pollInterval   time.Duration
	reportInterval time.Duration
	url            string
}

func (c *config) String() string {
	return fmt.Sprintf("config: pollInterval: %fs, reportInterval: %fs, url: \"%s\"",
		c.pollInterval.Seconds(),
		c.pollInterval.Seconds(),
		c.url)
}

func main() {
	log.Println("Agent started")

	cfg := &config{
		time.Second * 2,
		time.Second * 10,
		"http://127.0.0.1:8080",
	}

	log.Println(cfg)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	pollTicker := time.NewTicker(cfg.pollInterval)
	reportTicker := time.NewTicker(cfg.reportInterval)
	defer pollTicker.Stop()
	defer reportTicker.Stop()

	rand.Seed(time.Now().UnixNano())
	mc := collector.NewMetricCollector(cfg.url)
	for {
		select {
		case <-sigChan:
			log.Println("Agent successfully shutdown")
			return
		case <-pollTicker.C:
			mc.Collect()
		case <-reportTicker.C:
			mc.Report()
		}
	}
}
