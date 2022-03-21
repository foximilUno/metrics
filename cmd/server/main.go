package main

import (
	"encoding/json"
	"github.com/foximilUno/metrics/internal/config"
	"github.com/foximilUno/metrics/internal/repositories"
	"github.com/foximilUno/metrics/internal/server"
	st "github.com/foximilUno/metrics/internal/storage"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	cfg, err := config.InitMetricServerConfig()
	if err != nil {
		log.Fatalf("cant start server :%e", err)
	}

	if err := json.NewEncoder(log.Writer()).Encode(cfg); err != nil {
		log.Fatal("encoder err")
	}

	storage, err := st.NewMapStorage(cfg.DatabaseDsn)
	if err != nil {
		log.Fatalf("cant init storage: %e", err)
	}

	if len(cfg.DatabaseDsn) != 0 {
		if cfg.Restore {
			log.Printf("Restore from %s\r", cfg.DatabaseDsn)
			err := storage.Load()

			if err != nil {
				log.Printf("cant load from  %s: %e\n", cfg.DatabaseDsn, err)
			}
		}

		saveTicker := time.NewTicker(cfg.StoreInterval)

		go runTicker(saveTicker, storage)

	} else {
		log.Println("function \"Dump to file\" is turned off")
	}

	metricServer, err := server.NewMetricServer(cfg, storage)
	if err != nil {
		log.Fatalf("cant start metricServer: %e", err)
	}

	go metricServer.RunServer()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	<-sigChan
	log.Println("save on exit")
	if err := storage.Dump(); err != nil {
		log.Println(err)
		return
	}
}

func runTicker(ticker *time.Ticker, storage repositories.MetricSaver) {
	for {
		select {
		case <-ticker.C:
			if err := storage.Dump(); err != nil {
				log.Printf("cant save : err:%e", err)
			}
		default:
			time.Sleep(1 * time.Second)
		}
	}
}
