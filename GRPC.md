# gRPC API Documentation

This document describes the gRPC API for Promenitheus.

## Overview

Promenitheus provides a full-featured gRPC API alongside its HTTP API. The gRPC service runs on port 9091 (HTTP port + 1) by default.

## Service Definition

The gRPC service is defined in `api/proto/metrics.proto`:

```protobuf
service MetricsService {
  rpc GetMetrics(GetMetricsRequest) returns (GetMetricsResponse);
  rpc QueryMetrics(QueryMetricsRequest) returns (QueryMetricsResponse);
  rpc ListMetrics(ListMetricsRequest) returns (ListMetricsResponse);
}
```

## API Methods

### GetMetrics

Returns all metrics in Prometheus text format.

**Request**: `GetMetricsRequest` (empty)

**Response**: `GetMetricsResponse`
```json
{
  "content": "# TYPE metric_name counter\nmetric_name{label=\"value\"} 42\n",
  "contentType": "text/plain; version=0.0.4"
}
```

**Example**:
```bash
grpcurl -plaintext localhost:9091 promenitheus.v1.MetricsService/GetMetrics
```

### QueryMetrics

Query metrics by name, returning structured data.

**Request**: `QueryMetricsRequest`
```json
{
  "query": "metric_name"  // Optional: empty returns all
}
```

**Response**: `QueryMetricsResponse`
```json
{
  "status": "success",
  "data": [
    {
      "name": "metric_name",
      "type": "counter",
      "value": 42.0,
      "labels": {"label1": "value1"},
      "timestamp": 1766691830
    }
  ]
}
```

**Example**:
```bash
# Query specific metric
grpcurl -plaintext -d '{"query": "http_requests_total"}' \
  localhost:9091 promenitheus.v1.MetricsService/QueryMetrics

# Query all metrics
grpcurl -plaintext -d '{}' \
  localhost:9091 promenitheus.v1.MetricsService/QueryMetrics
```

### ListMetrics

List all metrics with optional filtering.

**Request**: `ListMetricsRequest`
```json
{
  "filter": "metric_name"  // Optional: filter by name
}
```

**Response**: `ListMetricsResponse`
```json
{
  "metrics": [
    {
      "name": "metric_name",
      "type": "gauge",
      "value": 3.14,
      "labels": {"env": "prod"},
      "timestamp": 1766691830
    }
  ]
}
```

**Example**:
```bash
# List all metrics
grpcurl -plaintext localhost:9091 promenitheus.v1.MetricsService/ListMetrics

# List with filter
grpcurl -plaintext -d '{"filter": "cpu_usage"}' \
  localhost:9091 promenitheus.v1.MetricsService/ListMetrics
```

## Message Types

### Metric

```protobuf
message Metric {
  string name = 1;                    // Metric name
  string type = 2;                    // "counter" or "gauge"
  double value = 3;                   // Current value
  map<string, string> labels = 4;     // Label key-value pairs
  int64 timestamp = 5;                // Unix timestamp (seconds)
}
```

## Service Discovery

Promenitheus includes gRPC reflection for easy service discovery:

```bash
# List all services
grpcurl -plaintext localhost:9091 list

# List service methods
grpcurl -plaintext localhost:9091 list promenitheus.v1.MetricsService

# Describe a method
grpcurl -plaintext localhost:9091 describe promenitheus.v1.MetricsService.QueryMetrics
```

## Client Examples

### Go Client

```go
import (
    pb "github.com/Avinash7390/Promenitheus/api/proto/v1"
    "google.golang.org/grpc"
)

conn, err := grpc.Dial("localhost:9091", grpc.WithInsecure())
if err != nil {
    log.Fatal(err)
}
defer conn.Close()

client := pb.NewMetricsServiceClient(conn)

// Query metrics
resp, err := client.QueryMetrics(context.Background(), &pb.QueryMetricsRequest{
    Query: "http_requests_total",
})
```

### Python Client

```python
import grpc
from api.proto.v1 import metrics_pb2, metrics_pb2_grpc

channel = grpc.insecure_channel('localhost:9091')
client = metrics_pb2_grpc.MetricsServiceStub(channel)

# Query metrics
response = client.QueryMetrics(
    metrics_pb2.QueryMetricsRequest(query='http_requests_total')
)

for metric in response.data:
    print(f"{metric.name}: {metric.value}")
```

## Advantages of gRPC

1. **Type Safety**: Protocol Buffers provide strong typing
2. **Performance**: Binary protocol is more efficient than JSON
3. **Streaming**: Ready for future streaming implementations
4. **Code Generation**: Auto-generate clients in multiple languages
5. **HTTP/2**: Built on HTTP/2 for better performance

## Regenerating Protocol Buffers

If you modify `api/proto/metrics.proto`, regenerate the code:

```bash
# Using Make
make proto

# Or manually
protoc --proto_path=api/proto \
  --go_out=api/proto/v1 --go_opt=paths=source_relative \
  --go-grpc_out=api/proto/v1 --go-grpc_opt=paths=source_relative \
  api/proto/metrics.proto
```

## Testing with grpcurl

Install grpcurl:
```bash
# macOS
brew install grpcurl

# Linux
wget https://github.com/fullstorydev/grpcurl/releases/download/v1.9.1/grpcurl_1.9.1_linux_x86_64.tar.gz
tar -xvf grpcurl_1.9.1_linux_x86_64.tar.gz
sudo mv grpcurl /usr/local/bin/
```

Basic usage:
```bash
# List services
grpcurl -plaintext localhost:9091 list

# Call a method
grpcurl -plaintext localhost:9091 promenitheus.v1.MetricsService/GetMetrics

# With request data
grpcurl -plaintext -d '{"query": "metric_name"}' \
  localhost:9091 promenitheus.v1.MetricsService/QueryMetrics
```
