package models

import "time"

type Observation struct {
	DeviceID           string    `json:"device_id"`
	TemperatureCelsius float64   `json:"temperature_celsius"`
	HumidityPercent    float64   `json:"humidity_percent"`
	PressureHpa        float64   `json:"pressure_hpa"`
	Timestamp          int64     `json:"timestamp"`
	TimestampISO       time.Time `json:"timestamp_iso"`
	LastReadMs         int64     `json:"last_read_ms"`
}
