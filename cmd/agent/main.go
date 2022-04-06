package main

import (
	"context"
	"encoding/json"
	"github.com/foximilUno/metrics/internal/collector"
	"github.com/foximilUno/metrics/internal/config"
	"log"
	"math/rand"
	"net/http"
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
	mc := collector.NewMetricCollector(cfg).WithClient(&http.Client{Timeout: 3 * time.Second})
	ctx, cancel := context.WithCancel(context.Background())

	collectCh := make(chan time.Time)
	collectAdditionalCh := make(chan time.Time)

	//fanout goroutine
	go func(ctx2 context.Context, ch <-chan time.Time) {
		for {
			select {
			case <-ctx.Done():
				log.Println(" shutdown broadcaster")
				return
			case v := <-ch:
				collectCh <- v
				collectAdditionalCh <- v
			}
		}
	}(ctx, pollTicker.C)

	//collect std goroutine
	go func(ctx2 context.Context, ch <-chan time.Time) {
		for {
			select {
			case <-ctx.Done():
				log.Println(" shutdown collect")
				return
			case <-ch:
				mc.Collect()
			}
		}
	}(ctx, collectCh)

	go func(ctx2 context.Context, ch <-chan time.Time) {
		for {
			select {
			case <-ctx.Done():
				log.Println(" shutdown collect additional")
				return
			case <-ch:
				mc.CollectAdditional()
			}
		}
	}(ctx, collectAdditionalCh)

	//report goroutine
	go func(ctx2 context.Context, ch <-chan time.Time) {
		for {
			select {
			case <-ctx.Done():
				log.Println(" shutdown report")
				return
			case <-ch:
				err = mc.Report()
				if err != nil {
					log.Printf("error while report: %s", err)
				}
			}
		}
	}(ctx, reportTicker.C)

	for range sigChan {
		log.Println("Agent successfully shutdown")
		cancel()
		return
	}
}
