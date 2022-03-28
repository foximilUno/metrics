package main

import (
	"encoding/json"
	"github.com/foximilUno/metrics/internal/config"
	"github.com/foximilUno/metrics/internal/repositories"
	"github.com/foximilUno/metrics/internal/server"
	st "github.com/foximilUno/metrics/internal/storage"
	"github.com/foximilUno/metrics/internal/storage/db"
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

	var storage repositories.MetricSaver
	var mapStorage *st.MapStorage

	storage, err = db.NewDbStorage(cfg.DatabaseDsn)
	if len(cfg.DatabaseDsn) != 0 {
		storage, err = db.NewDbStorage(cfg.DatabaseDsn)
	} else {
		mapStorage = st.NewMapStorage()
		err = mapStorage.Prepare(cfg)
		storage = repositories.MetricSaver(mapStorage)
		if err != nil {
			log.Fatalf("error while init map storage: %e", err)
		}
	}

	if err != nil {
		log.Fatalf("problem with establish connection to storage: %e", err)
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
	if len(cfg.StoreFile) != 0 {
		log.Println("save on exit")
		if mapStorage != nil {
			if err := mapStorage.Dump(); err != nil {
				log.Println(err)
				return
			}
		}
	}
}
