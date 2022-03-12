package server

import (
	"github.com/foximilUno/metrics/internal/config"
	"github.com/foximilUno/metrics/internal/handlers"
	"github.com/foximilUno/metrics/internal/repositories"
	"github.com/go-chi/chi"
	"log"
	"net/http"
)

type server struct {
	Server  *http.Server
	storage repositories.MetricSaver
}

func NewMetricServer(cfg *config.MetricServerConfig, storage repositories.MetricSaver) (*server, error) {
	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Route("/update", func(r chi.Router) {
			r.Post("/{metricType}/{metricName}/{metricVal}", handlers.SaveMetricsViaTextPlain(storage))
			r.Post("/", handlers.SaveMetricsViaJSON(storage))

		})

		r.Route("/value", func(r chi.Router) {
			r.Get("/{metricType}/{metricName}", handlers.GetMetricViaTextPlain(storage))
			r.Post("/", handlers.GetMetricViaJSON(storage))
		})
	})

	r.Get("/", handlers.GetMetricsTable(storage))

	curServer := &http.Server{
		Addr:    cfg.Host,
		Handler: r,
	}

	return &server{
		curServer,
		storage,
	}, nil
}

func (s *server) RunServer() {
	log.Printf("server started at endpoint %s\n", s.Server.Addr)
	if err := s.Server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
