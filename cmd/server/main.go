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

	storage := st.NewMapStorage()

	if len(cfg.DatabaseDsn) != 0 {
		persist, err := st.NewDbPersist(cfg.DatabaseDsn)
		if err != nil {
			log.Fatalf("problem with create db persist: %e", err)
		}
		err = storage.WithPersist(persist)
		if err != nil {
			log.Fatalf("cant init storage: %e", err)
		}

		if cfg.Restore {
			log.Printf("Restore from file %s\r", cfg.StoreFile)
			err := storage.Load()

			if err != nil {
				log.Printf("cant load from file %s: %e\n", cfg.StoreFile, err)
			}
		}

		saveTicker := time.NewTicker(cfg.StoreInterval)

		go runTicker(saveTicker, storage)

	} else {
		log.Println("function \"Dump\" is turned off")
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
	if storage.IsPersisted() {
		if err := storage.Dump(); err != nil {
			log.Println(err)
			return
		}
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
