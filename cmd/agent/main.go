package main

import (
	"fmt"
	"github.com/caarlos0/env"
	"github.com/foximilUno/metrics/internal/collector"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Config struct {
	PollInterval   time.Duration `env:"POLL_INTERVAL" envDefault:"2s"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL" envDefault:"10s"`
	URL            string        `env:"ADDRESS" envDefault:"http://127.0.0.1:8080"`
}

func (c *Config) String() string {
	return fmt.Sprintf("Config: PollInterval: %s, ReportInterval: %s, URL: \"%s\"",
		c.PollInterval,
		c.ReportInterval,
		c.URL)
}

func main() {
	//TODO debug
	fmt.Println(os.Environ())

	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatalf("cant start agent: %e", err)
	} else {
		log.Println("agent started")
	}

	log.Println(cfg.String())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	pollTicker := time.NewTicker(cfg.PollInterval)
	reportTicker := time.NewTicker(cfg.ReportInterval)
	defer pollTicker.Stop()
	defer reportTicker.Stop()

	rand.Seed(time.Now().UnixNano())
	mc := collector.NewMetricCollector(cfg.URL)
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
