package main

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env"
	"github.com/foximilUno/metrics/internal/collector"
	"github.com/foximilUno/metrics/internal/types"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var cfg types.Config

func init() {
	flag.StringVar(&cfg.URL, "a", "http://127.0.0.1:8080", "endpoint where metric send")
	flag.DurationVar(&cfg.PollInterval, "p", 2*time.Second, "in what time metric collect in host")
	flag.DurationVar(&cfg.ReportInterval, "r", 10*time.Second, "in what time metric push to server")

	flag.Parse()

	var cfgEnv types.Config

	if _, isPresent := os.LookupEnv("ADDRESS"); isPresent {
		cfg.URL = cfgEnv.URL
	}
	if _, isPresent := os.LookupEnv("POLL_INTERVAL"); isPresent {
		cfg.PollInterval = cfgEnv.PollInterval
	}
	if _, isPresent := os.LookupEnv("REPORT_INTERVAL"); isPresent {
		cfg.ReportInterval = cfgEnv.ReportInterval
	}
}

func main() {
	//TODO debug
	fmt.Println(os.Environ())

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
