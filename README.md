# Promenitheus

A basic Prometheus-like metric scraper written in Go. This project simulates core features of [Prometheus](https://prometheus.io/), including metric scraping, storage, and exposition.

## Features

- 🎯 **Metric Scraping**: Periodically scrapes metrics from configured HTTP endpoints
- 📊 **Metric Types**: Supports counters and gauges
- 🏷️ **Labels**: Full support for metric labels and label enrichment
- ⚙️ **Configuration**: YAML-based configuration similar to Prometheus
- 🚀 **Single Port Architecture**: HTTP and gRPC on the same port using cmux
- 🔀 **Connection Multiplexing**: Intelligent routing based on protocol (HTTP/1.1 vs HTTP/2)
- 🌐 **grpc-gateway**: Automatic HTTP/JSON to gRPC translation
- 🔍 **Query API**: Multiple API styles (REST, gRPC, JSON)
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
┌─────────────────────────────────────────┐
│  Unified Server (Single Port via cmux)  │
│  ┌─────────────────────────────────┐    │
│  │  TCP Listener (port 9090)       │    │
│  └──────────┬──────────────────────┘    │
│             │                            │
│    ┌────────┴────────┐                  │
│    │      cmux       │                  │
│    └────┬──────┬─────┘                  │
│         │      │                        │
│    HTTP/1.1  HTTP/2                     │
│         │      │                        │
│    ┌────▼──┐  ┌▼────────┐              │
│    │  HTTP │  │  gRPC   │              │
│    │Handler│  │ Server  │              │
│    └───┬───┘  └─────────┘              │
│        │                                │
│   ┌────▼─────────┐                     │
│   │grpc-gateway  │ (HTTP→gRPC)         │
│   └──────────────┘                     │
└─────────────────────────────────────────┘
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

3. View the scraped metrics (all on port 9090):

```bash
# HTTP/REST - Prometheus text format
curl http://localhost:9090/metrics

# HTTP/REST - JSON query API (via grpc-gateway)
curl "http://localhost:9090/api/v1/query?query=http_requests_total"

# HTTP/REST - List all metrics in JSON (via grpc-gateway)
curl http://localhost:9090/api/v1/metrics

# gRPC - List services (same port!)
grpcurl -plaintext localhost:9090 list

# gRPC - Get metrics in Prometheus format
grpcurl -plaintext localhost:9090 promenitheus.v1.MetricsService/GetMetrics

# gRPC - Query specific metrics
grpcurl -plaintext -d '{"query": "http_requests_total"}' \
  localhost:9090 promenitheus.v1.MetricsService/QueryMetrics

# gRPC - List all metrics
grpcurl -plaintext localhost:9090 promenitheus.v1.MetricsService/ListMetrics
```

**Note**: Both HTTP/1.1 and gRPC (HTTP/2) work on the **same port (9090)** thanks to connection multiplexing!
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

### Single Port Architecture

**All APIs are accessible on the same port (9090 by default)** using connection multiplexing:
- HTTP/1.1 requests → HTTP handlers & grpc-gateway
- HTTP/2 gRPC requests → gRPC server

### HTTP/REST API (HTTP/1.1)

- `GET /` - Home page with API documentation
- `GET /metrics` - All collected metrics in Prometheus text format (custom handler)
- `GET /api/v1/query?query=<metric_name>` - Query specific metrics (JSON via grpc-gateway)
- `GET /api/v1/metrics?filter=<metric_name>` - List all metrics (JSON via grpc-gateway)

### gRPC API (HTTP/2)

The gRPC service is defined in `api/proto/metrics.proto`. **Same port as HTTP!**

- **MetricsService.GetMetrics** - Returns all metrics in Prometheus text format
  ```bash
  grpcurl -plaintext localhost:9090 promenitheus.v1.MetricsService/GetMetrics
  ```

- **MetricsService.QueryMetrics** - Query metrics by name
  ```bash
  grpcurl -plaintext -d '{"query": "metric_name"}' \
    localhost:9090 promenitheus.v1.MetricsService/QueryMetrics
  ```

- **MetricsService.ListMetrics** - List all metrics with optional filter
  ```bash
  grpcurl -plaintext -d '{"filter": "metric_name"}' \
    localhost:9090 promenitheus.v1.MetricsService/ListMetrics
  ```

### How It Works

1. **cmux** (connection multiplexer) inspects incoming connections
2. HTTP/2 connections with gRPC content-type → routed to gRPC server
3. HTTP/1.x connections → routed to HTTP server
4. HTTP handlers use **grpc-gateway** to translate HTTP/JSON → gRPC calls
5. `/metrics` endpoint uses custom handler for native Prometheus text format

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

## Single Port Architecture with cmux

Promenitheus uses **connection multiplexing** to serve both HTTP and gRPC on the **same port**:

### How It Works

1. **Single TCP Listener**: One port (default 9090) handles all traffic
2. **cmux**: Inspects incoming connections and routes based on protocol:
   - HTTP/1.1 → HTTP handlers + grpc-gateway
   - HTTP/2 with gRPC content-type → gRPC server
3. **grpc-gateway**: Translates HTTP/JSON requests to gRPC calls
4. **Native Handlers**: Direct HTTP handlers for special cases (e.g., `/metrics` Prometheus format)

### Benefits

- **Simplified Deployment**: One port to configure and expose
- **Firewall Friendly**: Only need to open a single port
- **Protocol Flexibility**: Clients choose HTTP/REST or gRPC
- **Backward Compatible**: Existing HTTP clients work unchanged
- **Type Safety**: gRPC provides strong typing for supported languages

### Protocol Comparison

| Feature | HTTP/REST (via grpc-gateway) | gRPC (direct) |
|---------|------------------------------|---------------|
| Format | JSON | Protocol Buffers |
| Port | 9090 (shared) | 9090 (shared) |
| Protocol | HTTP/1.1 | HTTP/2 |
| Discovery | Documentation page | gRPC Reflection |
| Tools | curl, wget, browsers | grpcurl, gRPC clients |
| Performance | Good | Excellent |

### Testing Both Protocols

Install `grpcurl` for gRPC testing:
```bash
# macOS
brew install grpcurl

# Linux
wget https://github.com/fullstorydev/grpcurl/releases/download/v1.9.1/grpcurl_1.9.1_linux_x86_64.tar.gz
tar -xvf grpcurl_1.9.1_linux_x86_64.tar.gz
sudo mv grpcurl /usr/local/bin/

# Test HTTP/REST
curl http://localhost:9090/api/v1/metrics

# Test gRPC (same port!)
grpcurl -plaintext localhost:9090 list
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is open source and available for educational purposes.

## Acknowledgments

Inspired by [Prometheus](https://prometheus.io/) - The Cloud Native Computing Foundation project.
