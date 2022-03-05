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
	PollInterval   int64  `env:"POLL_INTERVAL" envDefault:"2"`
	ReportInterval int64  `env:"REPORT_INTERVAL" envDefault:"10"`
	URL            string `env:"ADDRESS" envDefault:"http://127.0.0.1:8080"`
}

func (c *Config) String() string {
	return fmt.Sprintf("Config: PollInterval: %ds, ReportInterval: %ds, URL: \"%s\"",
		c.PollInterval,
		c.ReportInterval,
		c.URL)
}

func main() {

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

	pollTicker := time.NewTicker(time.Duration(cfg.PollInterval) * time.Second)
	reportTicker := time.NewTicker(time.Duration(cfg.ReportInterval) * time.Second)
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
