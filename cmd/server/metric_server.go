package main

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"
)

const (
	defaultEndpoint        = ":8080"
	defaultApplicationType = "text/plain"
	servicePath            = "update"
)

var allowedTypes = map[string]string{
	"gauge":   "gauge",
	"counter": "counter",
}

type server struct {
	server *http.Server
}

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//check method only POST
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", 405)
		return
	}
	//check content type only defaultApplicationType
	if r.Header.Get("Content-type") != defaultApplicationType {
		w.Header().Add("Allowed", "text/plain")
		http.Error(w, "Allowed text/plain only", 415)
		return
	}

	//check elements in path
	segments := strings.Split(strings.TrimLeft(r.URL.Path, "/"), "/")
	if len(segments) != 4 {
		http.Error(w, "path must be pattern like /update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>", 400)
		return
	}
	//check first path segment
	if segments[0] != servicePath {
		http.Error(w, "path must be pattern like /update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>", 400)
		return
	}
	//check allowed types
	if _, ok := allowedTypes[segments[1]]; !ok {
		http.Error(w,
			fmt.Sprintf("Bad request: %s cant be, use %s", segments[1], reflect.ValueOf(allowedTypes).MapKeys()),
			400)
		return
	}
	log.Printf("invoked update metric with type \"%s\" witn name \"%s\" with value \"%s\"", segments[1], segments[2], segments[3])
	w.WriteHeader(200)
}

// if endpoint doesnt init - create server with default endpoint
func NewMetricServer(endpoint ...string) *server {
	var currentEndpoint string
	if len(endpoint) == 0 {
		currentEndpoint = defaultEndpoint
	} else {
		currentEndpoint = endpoint[0]
	}
	sh := http.HandlerFunc(ServeHTTP)
	return &server{
		&http.Server{
			Addr:    currentEndpoint,
			Handler: sh,
		},
	}
}

func (s *server) ListenAndServe() {
	log.Printf("Server started at endpoint %s\r\n", defaultEndpoint)
	if err := s.server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
