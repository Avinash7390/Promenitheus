# Promenitheus

A basic Prometheus-like metric scraper written in Go. This project simulates core features of [Prometheus](https://prometheus.io/), including metric scraping, storage, and exposition.

## Features

- 🎯 **Metric Scraping**: Periodically scrapes metrics from configured HTTP endpoints
- 📊 **Metric Types**: Supports counters and gauges
- 🏷️ **Labels**: Full support for metric labels and label enrichment
- ⚙️ **Configuration**: YAML-based configuration similar to Prometheus
- 🌐 **Dual Protocol Support**: Both HTTP and gRPC APIs
- 🔍 **Query API**: Simple JSON API for querying metrics
- 📦 **In-Memory Storage**: Fast in-memory metric registry
- 🔄 **gRPC Reflection**: Built-in reflection for easy service discovery

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
┌─────────────────────────────────┐
│  Dual Protocol Server           │
│  ├─ HTTP Server (port 9090)     │
│  │  └─ /metrics, /api/v1/*      │
│  └─ gRPC Server (port 9091)     │
│     └─ MetricsService            │
└─────────────────────────────────┘
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
# HTTP - Prometheus format
curl http://localhost:9090/metrics

# HTTP - JSON format (query API)
curl "http://localhost:9090/api/v1/query?query=http_requests_total"

# HTTP - List all metrics in JSON
curl http://localhost:9090/api/v1/metrics

# gRPC - List services
grpcurl -plaintext localhost:9091 list

# gRPC - Get metrics in Prometheus format
grpcurl -plaintext localhost:9091 promenitheus.v1.MetricsService/GetMetrics

# gRPC - Query specific metrics
grpcurl -plaintext -d '{"query": "http_requests_total"}' \
  localhost:9091 promenitheus.v1.MetricsService/QueryMetrics

# gRPC - List all metrics
grpcurl -plaintext localhost:9091 promenitheus.v1.MetricsService/ListMetrics
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

#### HTTP API (Port 9090 by default)

- `GET /` - Home page with API documentation
- `GET /metrics` - All collected metrics in Prometheus text format
- `GET /api/v1/query?query=<metric_name>` - Query specific metrics (JSON format)
- `GET /api/v1/metrics?filter=<metric_name>` - List all metrics in structured JSON format

#### gRPC API (Port 9091 by default)

The gRPC service is defined in `api/proto/metrics.proto`:

- **MetricsService.GetMetrics** - Returns all metrics in Prometheus text format
  ```bash
  grpcurl -plaintext localhost:9091 promenitheus.v1.MetricsService/GetMetrics
  ```

- **MetricsService.QueryMetrics** - Query metrics by name
  ```bash
  grpcurl -plaintext -d '{"query": "metric_name"}' \
    localhost:9091 promenitheus.v1.MetricsService/QueryMetrics
  ```

- **MetricsService.ListMetrics** - List all metrics with optional filter
  ```bash
  grpcurl -plaintext -d '{"filter": "metric_name"}' \
    localhost:9091 promenitheus.v1.MetricsService/ListMetrics
  ```

**Note**: Both HTTP and gRPC servers start automatically. gRPC uses port `HTTP_PORT + 1`.

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
├── api/
│   └── proto/
│       ├── v1/                 # Generated gRPC code
│       └── metrics.proto       # Protocol Buffer definitions
├── cmd/
│   ├── promenitheus/           # Main scraper application
│   └── example-target/         # Example target service
├── pkg/
│   ├── config/                 # Configuration loading
│   ├── metrics/                # Metric types and registry
│   ├── scraper/                # HTTP scraping logic
│   ├── storage/                # HTTP/gRPC server for exposing metrics
│   └── grpcserver/             # gRPC service implementation
└── config.yaml                 # Sample configuration
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

## gRPC and HTTP Support

Promenitheus provides **dual protocol support**, allowing you to use either HTTP or gRPC to access metrics:

### Why Both Protocols?

- **HTTP**: Easy to use with curl, web browsers, and standard HTTP clients
- **gRPC**: Type-safe, efficient binary protocol with built-in streaming support (future feature)
- **Flexibility**: Choose the protocol that best fits your use case

### Protocol Differences

| Feature | HTTP | gRPC |
|---------|------|------|
| Format | JSON / Prometheus Text | Protocol Buffers |
| Port | 9090 (configurable) | 9091 (HTTP port + 1) |
| Discovery | Documentation page | gRPC Reflection |
| Tools | curl, wget, browsers | grpcurl, grpc_cli |
| Streaming | Not supported | Ready for future implementation |

### Testing gRPC

Install `grpcurl` for testing:
```bash
# macOS
brew install grpcurl

# Linux
wget https://github.com/fullstorydev/grpcurl/releases/download/v1.9.1/grpcurl_1.9.1_linux_x86_64.tar.gz
tar -xvf grpcurl_1.9.1_linux_x86_64.tar.gz
sudo mv grpcurl /usr/local/bin/

# Test
grpcurl -plaintext localhost:9091 list
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is open source and available for educational purposes.

## Acknowledgments

Inspired by [Prometheus](https://prometheus.io/) - The Cloud Native Computing Foundation project.
