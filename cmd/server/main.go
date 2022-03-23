package main

import (
	"encoding/json"
	"github.com/foximilUno/metrics/internal/config"
	"github.com/foximilUno/metrics/internal/repositories"
	"github.com/foximilUno/metrics/internal/server"
	"github.com/foximilUno/metrics/internal/storage"
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

	st := storage.NewMapStorage()

	if len(cfg.DatabaseDsn) != 0 {

		err = st.WithPersist(cfg.DatabaseDsn)
		if err != nil {
			//TODO
			//log.Fatalf("cant init st: %e", err)
			log.Printf("cant init st: %e\r\n", err)
		}

		if cfg.Restore {

			log.Printf("Restore from %s\r", cfg.DatabaseDsn)
			err = st.Load()

			if err != nil {
				log.Printf("cant load from  %s: %e\n", cfg.DatabaseDsn, err)
			}
		} else {
			log.Println("Start server without restoring from persist")
		}

		saveTicker := time.NewTicker(cfg.StoreInterval)

		go runTicker(saveTicker, st)

	} else {
		log.Println("function \"Dump\" is turned off")
	}

	metricServer, err := server.NewMetricServer(cfg, st)
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
	if len(cfg.DatabaseDsn) != 0 {
		log.Println("save on exit")
		if err := st.Dump(); err != nil {
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
