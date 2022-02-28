package main

import (
	"github.com/foximilUno/metrics/internal/handlers"
	st "github.com/foximilUno/metrics/internal/storage"
	"github.com/go-chi/chi"
	"log"
	"net/http"
)

const (
	defaultEndpoint = ":8080"
)

func main() {
	storage := st.NewMapStorage()

	r := chi.NewRouter()

	r.Post("/update/{metricType}/{metricName}/{metricVal}", handlers.SaveMetricsViaTextPlain(storage))
	r.Get("/value/{metricType}/{metricName}", handlers.GetMetricViaTextPlain(storage))

	r.Post("/update/", handlers.SaveMetricsViaJSON(storage))
	r.Post("/value/", handlers.GetMetricViaJSON(storage))

	r.Get("/", handlers.GetMetricsTable(storage))
	server := &http.Server{
		Addr:    defaultEndpoint,
		Handler: r,
	}
	log.Printf("Server started at endpoint %s\r\n", defaultEndpoint)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}

}
