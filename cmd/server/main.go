package main

import (
	"github.com/foximilUno/metrics/internal/handlers"
	st "github.com/foximilUno/metrics/internal/storage"
	"log"
	"net/http"
)

const (
	defaultEndpoint = ":8080"
)

func main() {
	storage := st.NewMapStorage()

	server := &http.Server{
		Addr:    defaultEndpoint,
		Handler: handlers.SaveMetrics(storage),
	}
	log.Printf("Server started at endpoint %s\r\n", defaultEndpoint)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}

}
