package main

import (
	"github.com/caarlos0/env"
	"github.com/foximilUno/metrics/internal/handlers"
	st "github.com/foximilUno/metrics/internal/storage"
	"github.com/go-chi/chi"
	"log"
	"net/http"
	"reflect"
	"strings"
)

type Config struct {
	Host string `env:"ADDRESS" envDefault:":8080"`
}

func main() {
	var cfg Config

	trimEnv := func(v string) (interface{}, error) {
		return strings.TrimSpace(v), nil
	}

	err := env.ParseWithFuncs(&cfg, map[reflect.Type]env.ParserFunc{
		reflect.TypeOf(cfg.Host): trimEnv,
	})
	if err != nil {
		log.Fatalf("cant start server: %e", err)
	}

	storage := st.NewMapStorage()

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
	server := &http.Server{
		Addr:    ":" + strings.Split(cfg.Host, ":")[1],
		Handler: r,
	}
	log.Printf("Server started at endpoint %s\n", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}

}
