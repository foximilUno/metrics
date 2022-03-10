package server

import (
	"github.com/go-chi/chi"
	"net/http"
	"time"
)

type MetricServerConfig struct {
	Host          string        `json:"host" env:"ADDRESS" envDefault:":8080"`
	StoreInterval time.Duration `json:"storeInterval" env:"STORE_INTERVAL" envDefault:"300s"`
	StoreFile     string        `json:"storeFile" env:"STORE_FILE" envDefault:"/tmp/devops-metrics-db.json"`
	Restore       bool          `json:"isRestored" env:"RESTORE" envDefault:"true"`
}

type server struct {
	Server *http.Server
}

func NewMetricServer(cfg *MetricServerConfig, r *chi.Mux) (*server, error) {
	curServer := &http.Server{
		Addr:    cfg.Host,
		Handler: r,
	}

	return &server{
		curServer,
	}, nil
}
