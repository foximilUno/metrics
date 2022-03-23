package main

import (
	"encoding/json"
	"github.com/foximilUno/metrics/internal/collector"
	"github.com/foximilUno/metrics/internal/config"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg, err := config.InitConfig()
	if err != nil {
		log.Fatalf("cant start agent: %e", err)
	}

	//DEBUG
	log.Printf("ARGS: %v", os.Args)
	log.Printf("ENVS: %v", os.Environ())

	log.Println("agent started")

	if err := json.NewEncoder(log.Writer()).Encode(cfg); err != nil {
		log.Fatal("encoder err")
	}
	//log.Println(cfg.String())

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
	mc := collector.NewMetricCollector(cfg)
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
