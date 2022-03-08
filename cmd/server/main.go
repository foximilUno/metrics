package main

import (
	"encoding/json"
	"github.com/caarlos0/env"
	"github.com/foximilUno/metrics/internal/handlers"
	"github.com/foximilUno/metrics/internal/repositories"
	"github.com/foximilUno/metrics/internal/server"
	st "github.com/foximilUno/metrics/internal/storage"
	"github.com/go-chi/chi"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	var cfg server.MetricServerConfig

	err := env.Parse(&cfg)
	if err != nil {
		log.Fatalf("cant load metricServer envs: %e", err)
	}
	err = json.NewEncoder(log.Writer()).Encode(cfg)
	if err != nil {
		return
	}
	storage := st.NewMapStorage()

	if len(cfg.StoreFile) != 0 {
		if cfg.Restore {
			log.Printf("Restore from file %s\r", cfg.StoreFile)
			err := storage.LoadFromFile(cfg.StoreFile)

			if err != nil {
				log.Fatalf("cant load from file %s: %e\n", cfg.StoreFile, err)
			}
		}

		saveTicker := time.NewTicker(time.Duration(cfg.StoreInterval) * time.Second)

		go dumpToFile(saveTicker, storage, cfg.StoreFile)

	} else {
		log.Println("function \"Save to file\" is turned off")

	}

	go runServer(cfg, storage)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	<-sigChan
	log.Println("save on exit")
	err = storage.SaveToFile(cfg.StoreFile)
	if err != nil {
		log.Println(err)
		return
	}
}

func dumpToFile(ticker *time.Ticker, storage repositories.MetricSaver, filepath string) {
	for {
		select {
		case <-ticker.C:
			if err := storage.SaveToFile(filepath); err != nil {
				log.Fatalf("cant save to file\"%s\", err:%e", filepath, err)
			}
		default:
			time.Sleep(1 * time.Second)
		}
	}
}

func runServer(cfg server.MetricServerConfig, storage repositories.MetricSaver) {
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

	metricServer, err := server.NewMetricServer(&cfg, r)

	if err != nil {
		log.Fatalf("cant start server: %e", err)
	}
	log.Printf("server started at endpoint %s\n", metricServer.Server.Addr)
	err = metricServer.Server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
