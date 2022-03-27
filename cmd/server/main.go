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

	err = storage.Prepare(cfg)
	if err != nil {
		log.Fatalf("problem with prepare storage: %e", err)
	}

	metricServer, err := server.NewMetricServer(cfg, repositories.MetricSaver(storage))
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
