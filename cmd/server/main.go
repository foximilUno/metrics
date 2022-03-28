package main

import (
	"encoding/json"
	"github.com/foximilUno/metrics/internal/config"
	"github.com/foximilUno/metrics/internal/repositories"
	"github.com/foximilUno/metrics/internal/server"
	"github.com/foximilUno/metrics/internal/storage/db"
	"log"
)

func main() {

	cfg, err := config.InitMetricServerConfig()
	if err != nil {
		log.Fatalf("cant start server :%e", err)
	}

	if err := json.NewEncoder(log.Writer()).Encode(cfg); err != nil {
		log.Fatal("encoder err")
	}

	storage, err := db.NewDbStorage(cfg.DatabaseDsn)

	if err != nil {
		log.Fatalf("problem with establish connection to storage: %e", err)
	}

	metricServer, err := server.NewMetricServer(cfg, repositories.MetricSaver(storage))
	if err != nil {
		log.Fatalf("cant start metricServer: %e", err)
	}

	metricServer.RunServer()
}
