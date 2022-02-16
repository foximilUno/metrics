package server

import (
	"github.com/foximilUno/metrics/internal/handlers"
	"log"
	"net/http"
)

const (
	defaultEndpoint = ":8080"
)

type server struct {
	server *http.Server
}

// if endpoint doesnt init - create server with default endpoint
func NewMetricServer(endpoint ...string) *server {
	var currentEndpoint string
	if len(endpoint) == 0 {
		currentEndpoint = defaultEndpoint
	} else {
		currentEndpoint = endpoint[0]
	}
	return &server{
		&http.Server{
			Addr:    currentEndpoint,
			Handler: handlers.SaveMetrics(),
		},
	}
}

func (s *server) ListenAndServe() {
	log.Printf("Server started at endpoint %s\r\n", defaultEndpoint)
	if err := s.server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
