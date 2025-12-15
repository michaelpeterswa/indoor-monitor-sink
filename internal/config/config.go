package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	LogLevel string `env:"LOG_LEVEL" envDefault:"error"`

	MetricsEnabled bool `env:"METRICS_ENABLED" envDefault:"true"`
	MetricsPort    int  `env:"METRICS_PORT" envDefault:"8081"`

	Local bool `env:"LOCAL" envDefault:"false"`

	TracingEnabled    bool    `env:"TRACING_ENABLED" envDefault:"false"`
	TracingSampleRate float64 `env:"TRACING_SAMPLERATE" envDefault:"0.01"`
	TracingService    string  `env:"TRACING_SERVICE" envDefault:"katalog-agent"`
	TracingVersion    string  `env:"TRACING_VERSION"`

	// Indoor Monitor Hosts (comma-separated list of host URLs)
	Hosts         []string `env:"HOSTS" envSeparator:","`
	PollInterval  int      `env:"POLL_INTERVAL" envDefault:"60"` // seconds

	// InfluxDB Configuration
	InfluxURL    string `env:"INFLUX_URL"`
	InfluxToken  string `env:"INFLUX_TOKEN"`
	InfluxOrg    string `env:"INFLUX_ORG"`
	InfluxBucket string `env:"INFLUX_BUCKET"`
}

func NewConfig() (*Config, error) {
	var cfg Config

	err := env.Parse(&cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &cfg, nil
}
