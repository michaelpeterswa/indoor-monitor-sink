package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"alpineworks.io/ootel"
	"github.com/michaelpeterswa/indoor-monitor-sink/internal/config"
	"github.com/michaelpeterswa/indoor-monitor-sink/internal/fetcher"
	"github.com/michaelpeterswa/indoor-monitor-sink/internal/logging"
	"github.com/michaelpeterswa/indoor-monitor-sink/internal/writer"
	"go.opentelemetry.io/contrib/instrumentation/host"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
)

func main() {
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "error"
	}

	slogLevel, err := logging.LogLevelToSlogLevel(logLevel)
	if err != nil {
		log.Fatalf("could not convert log level: %s", err)
	}

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slogLevel,
	})))
	c, err := config.NewConfig()
	if err != nil {
		slog.Error("could not create config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	ctx := context.Background()

	exporterType := ootel.ExporterTypePrometheus
	if c.Local {
		exporterType = ootel.ExporterTypeOTLPGRPC
	}

	ootelClient := ootel.NewOotelClient(
		ootel.WithMetricConfig(
			ootel.NewMetricConfig(
				c.MetricsEnabled,
				exporterType,
				c.MetricsPort,
			),
		),
		ootel.WithTraceConfig(
			ootel.NewTraceConfig(
				c.TracingEnabled,
				c.TracingSampleRate,
				c.TracingService,
				c.TracingVersion,
			),
		),
	)

	shutdown, err := ootelClient.Init(ctx)
	if err != nil {
		slog.Error("could not create ootel client", slog.String("error", err.Error()))
		os.Exit(1)
	}

	err = runtime.Start(runtime.WithMinimumReadMemStatsInterval(5 * time.Second))
	if err != nil {
		slog.Error("could not create runtime metrics", slog.String("error", err.Error()))
		os.Exit(1)
	}

	err = host.Start()
	if err != nil {
		slog.Error("could not create host metrics", slog.String("error", err.Error()))
		os.Exit(1)
	}

	defer func() {
		_ = shutdown(ctx)
	}()

	// Validate configuration
	if len(c.Hosts) == 0 {
		slog.Error("no hosts configured, set HOSTS environment variable")
		os.Exit(1)
	}

	if c.InfluxURL == "" || c.InfluxToken == "" || c.InfluxOrg == "" || c.InfluxBucket == "" {
		slog.Error("influxdb configuration incomplete, check INFLUX_URL, INFLUX_TOKEN, INFLUX_ORG, INFLUX_BUCKET")
		os.Exit(1)
	}

	// Initialize services
	f := fetcher.NewFetcher(c.Hosts)
	influxWriter, err := writer.NewInfluxWriter(c.InfluxURL, c.InfluxToken, c.InfluxOrg, c.InfluxBucket)
	if err != nil {
		slog.Error("could not create influx writer", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer influxWriter.Close()

	slog.Info("indoor-monitor-sink started",
		slog.Int("host_count", len(c.Hosts)),
		slog.Int("poll_interval_seconds", c.PollInterval),
	)

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	ticker := time.NewTicker(time.Duration(c.PollInterval) * time.Second)
	defer ticker.Stop()

	// Do initial fetch
	observations := f.FetchAll(ctx)
	if len(observations) > 0 {
		if err := influxWriter.WriteObservations(ctx, observations); err != nil {
			slog.Error("failed to write observations", slog.String("error", err.Error()))
		}
	}

	// Main loop
	for {
		select {
		case <-ticker.C:
			observations := f.FetchAll(ctx)
			if len(observations) > 0 {
				if err := influxWriter.WriteObservations(ctx, observations); err != nil {
					slog.Error("failed to write observations", slog.String("error", err.Error()))
				}
			}
		case sig := <-sigChan:
			slog.Info("received signal, shutting down", slog.String("signal", sig.String()))
			return
		}
	}
}
