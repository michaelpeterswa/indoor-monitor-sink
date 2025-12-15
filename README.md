# indoor-monitor-sink

A Go application that periodically polls indoor monitoring devices for environmental observations and writes them to InfluxDB.

## Features

- Polls multiple host endpoints for temperature, humidity, and pressure observations
- Writes data to InfluxDB with proper timestamping
- Configurable polling interval
- Graceful shutdown handling
- OpenTelemetry metrics and tracing support
- Structured JSON logging

## Environment Variables

### Required

- `HOSTS` - Comma-separated list of host URLs to poll (e.g., `http://device1.local,http://device2.local`)
- `INFLUX_URL` - InfluxDB server URL (e.g., `http://localhost:8086`)
- `INFLUX_TOKEN` - InfluxDB authentication token
- `INFLUX_ORG` - InfluxDB organization name
- `INFLUX_BUCKET` - InfluxDB bucket name

### Optional

- `POLL_INTERVAL` - Polling interval in seconds (default: `60`)
- `LOG_LEVEL` - Log level: `debug`, `info`, `warn`, `error` (default: `error`)
- `METRICS_ENABLED` - Enable Prometheus metrics (default: `true`)
- `METRICS_PORT` - Metrics server port (default: `8081`)
- `TRACING_ENABLED` - Enable OpenTelemetry tracing (default: `false`)
- `TRACING_SAMPLERATE` - Trace sampling rate (default: `0.01`)
- `TRACING_SERVICE` - Service name for tracing (default: `katalog-agent`)
- `LOCAL` - Use OTLP gRPC exporter instead of Prometheus (default: `false`)

## API Endpoint Format

The application expects each host to provide a GET endpoint at `/api/v1/observation` that returns JSON in this format:

```json
{
  "device_id": "nozlis",
  "temperature_celsius": 20.38,
  "humidity_percent": 41.84082,
  "pressure_hpa": 901.8663,
  "timestamp": 1765755580,
  "timestamp_iso": "2025-12-14T23:39:40Z",
  "last_read_ms": 11098025
}
```

## Usage

```bash
export HOSTS="http://device1.local,http://device2.local"
export INFLUX_URL="http://localhost:8086"
export INFLUX_TOKEN="your-token"
export INFLUX_ORG="your-org"
export INFLUX_BUCKET="indoor-monitoring"
export POLL_INTERVAL=30
export LOG_LEVEL="info"

./indoor-monitor-sink
```

## Building

```bash
go build ./cmd/indoor-monitor-sink
```

## InfluxDB Data Schema

Data is written to InfluxDB with the following schema:

- **Measurement**: `observation`
- **Tags**:
  - `device_id` - Unique identifier for the device
- **Fields**:
  - `temperature_celsius` - Temperature in Celsius
  - `humidity_percent` - Relative humidity percentage
  - `pressure_hpa` - Atmospheric pressure in hPa
  - `last_read_ms` - Time since last sensor read in milliseconds
- **Timestamp**: Unix timestamp from the observation
