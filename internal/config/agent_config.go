package config

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env"
	"os"
	"time"
)

type Config struct {
	PollInterval   time.Duration `env:"POLL_INTERVAL"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL"`
	URL            string        `env:"ADDRESS"`
	Key            string        `env:"KEY"`
}

func (c *Config) String() string {
	return fmt.Sprintf("Config: PollInterval: %s, ReportInterval: %s, URL: \"%s\"",
		c.PollInterval,
		c.ReportInterval,
		c.URL)
}

func InitConfig() (*Config, error) {
	var cfg Config
	flag.StringVar(&cfg.URL, "a", "http://127.0.0.1:8080", "endpoint where metric send")
	flag.DurationVar(&cfg.PollInterval, "p", 2*time.Second, "in what time metric collect in host")
	flag.DurationVar(&cfg.ReportInterval, "r", 10*time.Second, "in what time metric push to server")
	flag.StringVar(&cfg.Key, "k", "", "private key for hashing sha256")

	flag.Parse()

	var cfgEnv Config

	if err := env.Parse(&cfgEnv); err != nil {
		return nil, err
	}

	if _, isPresent := os.LookupEnv("ADDRESS"); isPresent {
		cfg.URL = cfgEnv.URL
	}
	if _, isPresent := os.LookupEnv("POLL_INTERVAL"); isPresent {
		cfg.PollInterval = cfgEnv.PollInterval
	}
	if _, isPresent := os.LookupEnv("REPORT_INTERVAL"); isPresent {
		cfg.ReportInterval = cfgEnv.ReportInterval
	}
	if _, isPresent := os.LookupEnv("KEY"); isPresent {
		cfg.Key = cfgEnv.Key
	}
	return &cfg, nil
}
