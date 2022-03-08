package server

import (
	"github.com/go-chi/chi"
	"net/http"
)

type MetricServerConfig struct {
	Host          string `env:"ADDRESS" envDefault:":8080"`
	StoreInterval int    `env:"STORE_INTERVAL" envDefault:"3"`
	StoreFile     string `env:"STORE_FILE" envDefault:"./devops-metrics-db.json"`
	Restore       bool   `env:"RESTORE" envDefault:"true"`
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
