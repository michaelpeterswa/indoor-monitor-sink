package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/michaelpeterswa/indoor-monitor-sink/internal/models"
)

type Fetcher struct {
	hosts      []string
	httpClient *http.Client
}

func NewFetcher(hosts []string) *Fetcher {
	return &Fetcher{
		hosts: hosts,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (f *Fetcher) FetchObservation(ctx context.Context, host string) (*models.Observation, error) {
	url := fmt.Sprintf("%s/api/v1/observation", host)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch observation: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	var obs models.Observation
	if err := json.NewDecoder(resp.Body).Decode(&obs); err != nil {
		return nil, fmt.Errorf("failed to decode observation: %w", err)
	}

	return &obs, nil
}

func (f *Fetcher) FetchAll(ctx context.Context) []*models.Observation {
	observations := make([]*models.Observation, 0, len(f.hosts))

	for _, host := range f.hosts {
		obs, err := f.FetchObservation(ctx, host)
		if err != nil {
			slog.Error("failed to fetch observation",
				slog.String("host", host),
				slog.String("error", err.Error()),
			)
			continue
		}

		slog.Info("fetched observation",
			slog.String("host", host),
			slog.String("device_id", obs.DeviceID),
			slog.Float64("temperature_celsius", obs.TemperatureCelsius),
		)

		observations = append(observations, obs)
	}

	return observations
}
