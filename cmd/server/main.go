package main

import (
	"encoding/json"
	"flag"
	"github.com/caarlos0/env"
	"github.com/foximilUno/metrics/internal/repositories"
	"github.com/foximilUno/metrics/internal/server"
	st "github.com/foximilUno/metrics/internal/storage"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var cfg server.MetricServerConfig

//init server config
func init() {
	flag.StringVar(&cfg.Host, "a", "localhost:8080", "server url as <host:port>")
	flag.BoolVar(&cfg.Restore, "r", true, "is restored from file - <true/false>")
	flag.StringVar(&cfg.StoreFile, "f", "/tmp/devops-metrics-db.json", "path to file to load/save metrics")
	flag.DurationVar(&cfg.StoreInterval, "i", time.Duration(300*time.Second), "with interval save to file")

	flag.Parse()

	var cfgEnv server.MetricServerConfig

	if err := env.Parse(&cfgEnv); err != nil {
		log.Fatalf("cant load metricServer envs: %e", err)
	}

	if len(cfgEnv.Host) != 0 {
		cfg.Host = cfgEnv.Host
	}
	if len(cfgEnv.StoreFile) != 0 {
		cfg.StoreFile = cfgEnv.StoreFile
	}
	if len(os.Getenv("RESTORE")) != 0 {
		cfg.Restore = cfgEnv.Restore
	}
	if len(os.Getenv("STORE_INTERVAL")) != 0 {
		cfg.StoreInterval = cfgEnv.StoreInterval
	}

}

func main() {
	if err := json.NewEncoder(log.Writer()).Encode(cfg); err != nil {
		return
	}

	storage := st.NewMapStorage()

	if len(cfg.StoreFile) != 0 {
		if cfg.Restore {
			log.Printf("Restore from file %s\r", cfg.StoreFile)
			err := storage.LoadFromFile(cfg.StoreFile)

			if err != nil {
				log.Printf("cant load from file %s: %e\n", cfg.StoreFile, err)
			}
		}

		saveTicker := time.NewTicker(cfg.StoreInterval)

		go func(ticker *time.Ticker, storage repositories.MetricSaver, filepath string) {
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
		}(saveTicker, storage, cfg.StoreFile)

	} else {
		log.Println("function \"Save to file\" is turned off")
	}

	metricServer, err := server.NewMetricServer(cfg, storage)
	if err != nil {
		log.Fatalf("cant start metricServer: %e", err)
	}

	go metricServer.RunServer()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	<-sigChan
	log.Println("save on exit")
	if err := storage.SaveToFile(cfg.StoreFile); err != nil {
		log.Println(err)
		return
	}
}
