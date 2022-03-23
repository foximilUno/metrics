package config

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env"
	"os"
	"time"
)

type MetricServerConfig struct {
	Host          string        `json:"host" env:"ADDRESS"`
	StoreInterval time.Duration `json:"storeInterval" env:"STORE_INTERVAL"`
	//StoreFile     string        `json:"storeFile" env:"STORE_FILE"`
	Restore     bool   `json:"isRestored" env:"RESTORE"`
	Key         string `env:"KEY"`
	DatabaseDsn string `env:"DATABASE_DSN"`
}

func InitMetricServerConfig() (*MetricServerConfig, error) {
	var cfg MetricServerConfig

	flag.StringVar(&cfg.Host, "a", "localhost:8080", "server url as <host:port>")
	flag.BoolVar(&cfg.Restore, "r", true, "is restored from file - <true/false>")
	//flag.StringVar(&cfg.StoreFile, "f", "/tmp/devops-metrics-db.json", "path to file to load/save metrics")
	flag.DurationVar(&cfg.StoreInterval, "i", 300*time.Second, "with interval save to file")
	flag.StringVar(&cfg.Key, "k", "", "private key to check data incoming")
	flag.StringVar(&cfg.DatabaseDsn, "d", "", "database connection string")

	var cfgEnv MetricServerConfig

	if err := env.Parse(&cfgEnv); err != nil {
		return nil, fmt.Errorf("cant load metricServer envs: %e", err)
	}

	//DEBUG change flag.Parse() to inner invoke to catch error
	//flag.Parse()
	if err := flag.CommandLine.Parse(os.Args[1:]); err != nil {
		return nil, err
	}

	if len(cfgEnv.Host) != 0 {
		cfg.Host = cfgEnv.Host
	}
	if len(os.Getenv("RESTORE")) != 0 {
		cfg.Restore = cfgEnv.Restore
	}
	if len(os.Getenv("STORE_INTERVAL")) != 0 {
		cfg.StoreInterval = cfgEnv.StoreInterval
	}
	if _, isPresent := os.LookupEnv("KEY"); isPresent {
		cfg.Key = cfgEnv.Key
	}
	if _, isPresent := os.LookupEnv("DATABASE_DSN"); isPresent {
		cfg.DatabaseDsn = cfgEnv.DatabaseDsn
	}

	return &cfg, nil
}
