package main

import (
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

	log.Println(cfg)

	storage := st.NewMapStorage()

	if len(cfg.StoreFile) != 0 {
		if cfg.Restore {
			log.Printf("Restore from file %s\r", cfg.StoreFile)
			err := storage.LoadFromFile(cfg.StoreFile)

			if err != nil {
				log.Printf("cant load from file %s: %e\n", cfg.StoreFile, err)
			}
		}

		saveTicker := time.NewTicker(cfg.StoreInterval)

		go func(ticker *time.Ticker, storage repositories.MetricSaver, filepath string) {
			for {
				select {
				case <-ticker.C:
					if err := storage.SaveToFile(filepath); err != nil {
						log.Printf("cant save to file\"%s\", err:%e", filepath, err)
					}
				default:
					time.Sleep(1 * time.Second)
				}
			}
		}(saveTicker, storage, cfg.StoreFile)

	} else {
		log.Println("function \"Save to file\" is turned off")
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
	if err := storage.SaveToFile(cfg.StoreFile); err != nil {
		log.Println(err)
		return
	}
}
