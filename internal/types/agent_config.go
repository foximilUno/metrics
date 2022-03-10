package types

import (
	"fmt"
	"time"
)

type Config struct {
	PollInterval   time.Duration `env:"POLL_INTERVAL"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL"`
	URL            string        `env:"ADDRESS"`
}

func (c *Config) String() string {
	return fmt.Sprintf("Config: PollInterval: %s, ReportInterval: %s, URL: \"%s\"",
		c.PollInterval,
		c.ReportInterval,
		c.URL)
}
