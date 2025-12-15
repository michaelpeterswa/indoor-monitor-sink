package writer

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/michaelpeterswa/indoor-monitor-sink/internal/models"
)

type InfluxWriter struct {
	client influxdb2.Client
	org    string
	bucket string
}

func NewInfluxWriter(url, token, org, bucket string) (*InfluxWriter, error) {
	client := influxdb2.NewClient(url, token)

	return &InfluxWriter{
		client: client,
		org:    org,
		bucket: bucket,
	}, nil
}

func (w *InfluxWriter) WriteObservations(ctx context.Context, observations []*models.Observation) error {
	writeAPI := w.client.WriteAPIBlocking(w.org, w.bucket)

	points := make([]*write.Point, 0, len(observations))

	for _, obs := range observations {
		timestamp := time.Unix(obs.Timestamp, 0)

		point := influxdb2.NewPoint(
			"observation",
			map[string]string{
				"device_id": obs.DeviceID,
			},
			map[string]interface{}{
				"temperature_celsius": obs.TemperatureCelsius,
				"humidity_percent":    obs.HumidityPercent,
				"pressure_hpa":        obs.PressureHpa,
				"last_read_ms":        obs.LastReadMs,
			},
			timestamp,
		)

		points = append(points, point)

		slog.Debug("prepared influx point",
			slog.String("device_id", obs.DeviceID),
			slog.Time("timestamp", timestamp),
		)
	}

	if len(points) == 0 {
		slog.Warn("no points to write to influx")
		return nil
	}

	if err := writeAPI.WritePoint(ctx, points...); err != nil {
		return fmt.Errorf("failed to write to influx: %w", err)
	}

	slog.Info("wrote observations to influx", slog.Int("count", len(points)))

	return nil
}

func (w *InfluxWriter) Close() {
	w.client.Close()
}
