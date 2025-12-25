# Promenitheus

A basic Prometheus-like metric scraper written in Go. This project simulates core features of [Prometheus](https://prometheus.io/), including metric scraping, storage, and exposition.

## Features

- 🎯 **Metric Scraping**: Periodically scrapes metrics from configured HTTP endpoints
- 📊 **Metric Types**: Supports counters and gauges
- 🏷️ **Labels**: Full support for metric labels and label enrichment
- ⚙️ **Configuration**: YAML-based configuration similar to Prometheus
- 🌐 **HTTP API**: Exposes collected metrics in Prometheus text format
- 🔍 **Query API**: Simple JSON API for querying metrics
- 📦 **In-Memory Storage**: Fast in-memory metric registry

## Architecture

```
┌─────────────────┐
│  Target Service │ (exposes /metrics)
└────────┬────────┘
         │ HTTP GET
         ▼
┌─────────────────┐
│    Scraper      │ (collects metrics periodically)
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Metric Registry │ (stores metrics in memory)
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  HTTP Server    │ (exposes /metrics and /api/v1/query)
└─────────────────┘
```

## Getting Started

### Prerequisites

- Go 1.21 or higher

### Installation

Clone the repository and build the binaries:

```bash
git clone https://github.com/Avinash7390/Promenitheus.git
cd Promenitheus

# Using Make (recommended)
make build

# Or manually
go build -o bin/promenitheus ./cmd/promenitheus
go build -o bin/example-target ./cmd/example-target
```

### Running the Example

1. Start the example target service (exposes sample metrics):

```bash
# Using Make
make run-target

# Or manually
./bin/example-target --port 8080
```

2. In another terminal, start Promenitheus with the sample configuration:

```bash
# Using Make
make run-scraper

# Or manually
./bin/promenitheus --config config.yaml --port 9090
```

3. View the scraped metrics:

```bash
# Prometheus format
curl http://localhost:9090/metrics

# JSON format (query API)
curl "http://localhost:9090/api/v1/query?query=http_requests_total"
```

## Configuration

Create a `config.yaml` file with your scrape targets:

```yaml
global:
  scrape_interval: 15s  # How often to scrape targets
  scrape_timeout: 10s   # Timeout for scrape requests

scrape_configs:
  - job_name: 'my-service'
    scrape_interval: 10s  # Override global interval
    static_configs:
      - targets:
          - 'localhost:8080'
        labels:
          environment: 'production'
          region: 'us-west'
```

### Configuration Options

- `global.scrape_interval`: Default interval between scrapes (default: 15s)
- `global.scrape_timeout`: Default timeout for scrape requests (default: 10s)
- `scrape_configs[].job_name`: Name of the scrape job (added as `job` label)
- `scrape_configs[].scrape_interval`: Per-job scrape interval (overrides global)
- `scrape_configs[].static_configs[].targets`: List of `host:port` targets to scrape
- `scrape_configs[].static_configs[].labels`: Additional labels to add to scraped metrics

## API Endpoints

### Promenitheus Server

- `GET /` - Home page with links
- `GET /metrics` - All collected metrics in Prometheus text format
- `GET /api/v1/query?query=<metric_name>` - Query specific metrics (JSON format)

### Example Target Service

- `GET /` - Home page
- `GET /metrics` - Exposed metrics in Prometheus format

## Metric Format

Promenitheus uses the Prometheus text exposition format:

```
# TYPE metric_name metric_type
metric_name{label1="value1",label2="value2"} 42
```

Example:
```
# TYPE http_requests_total counter
http_requests_total{method="GET",endpoint="/api"} 1234
http_requests_total{method="POST",endpoint="/api"} 567
```

## Development

### Project Structure

```
.
├── cmd/
│   ├── promenitheus/       # Main scraper application
│   └── example-target/     # Example target service
├── pkg/
│   ├── config/             # Configuration loading
│   ├── metrics/            # Metric types and registry
│   ├── scraper/            # HTTP scraping logic
│   └── storage/            # HTTP server for exposing metrics
└── config.yaml             # Sample configuration
```

### Running Tests

```bash
go test ./... -v
```

### Building

```bash
# Build both binaries using Make
make build

# Run tests
make test

# Clean build artifacts
make clean

# See all available targets
make help

# Or build manually
go build -o bin/promenitheus ./cmd/promenitheus
go build -o bin/example-target ./cmd/example-target

# Or build everything
go build ./...
```

## Examples

### Exposing Custom Metrics

To make your application scrapeable by Promenitheus, expose a `/metrics` endpoint:

```go
func metricsHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/plain; version=0.0.4")
    
    fmt.Fprintf(w, "# TYPE my_counter counter\n")
    fmt.Fprintf(w, "my_counter 42\n")
    
    fmt.Fprintf(w, "# TYPE my_gauge gauge\n")
    fmt.Fprintf(w, "my_gauge 3.14\n")
}
```

## Differences from Prometheus

This is a simplified implementation for educational purposes. Notable differences:

- **Storage**: In-memory only (no persistent storage)
- **Query Language**: Simple metric name queries only (no PromQL)
- **Metric Types**: Only counters and gauges (no histograms or summaries)
- **Service Discovery**: Static configuration only
- **Alerting**: Not implemented
- **Recording Rules**: Not implemented

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is open source and available for educational purposes.

## Acknowledgments

Inspired by [Prometheus](https://prometheus.io/) - The Cloud Native Computing Foundation project.
