package grpcserver

import (
	"context"
	"strings"
	"testing"
	"time"

	pb "github.com/Avinash7390/Promenitheus/api/proto/v1"
	"github.com/Avinash7390/Promenitheus/pkg/metrics"
)

func TestMetricsServer(t *testing.T) {
	registry := metrics.NewMetricRegistry()
	server := NewMetricsServer(registry)

	t.Run("GetMetrics returns Prometheus format", func(t *testing.T) {
		// Add some test metrics
		registry.Register(&metrics.Metric{
			Name:  "test_counter",
			Type:  metrics.MetricTypeCounter,
			Value: 42.0,
			Labels: map[string]string{
				"label1": "value1",
			},
		})

		resp, err := server.GetMetrics(context.Background(), &pb.GetMetricsRequest{})
		if err != nil {
			t.Fatalf("GetMetrics failed: %v", err)
		}

		if resp.ContentType != "text/plain; version=0.0.4" {
			t.Errorf("Expected content type 'text/plain; version=0.0.4', got '%s'", resp.ContentType)
		}

		if !strings.Contains(resp.Content, "test_counter") {
			t.Error("Response should contain test_counter")
		}

		if !strings.Contains(resp.Content, "# TYPE test_counter counter") {
			t.Error("Response should contain TYPE hint")
		}

		if !strings.Contains(resp.Content, `label1="value1"`) {
			t.Error("Response should contain labels")
		}
	})

	t.Run("QueryMetrics filters by name", func(t *testing.T) {
		registry.Clear()

		registry.Register(&metrics.Metric{
			Name:  "metric1",
			Type:  metrics.MetricTypeCounter,
			Value: 10.0,
		})

		registry.Register(&metrics.Metric{
			Name:  "metric2",
			Type:  metrics.MetricTypeGauge,
			Value: 20.0,
		})

		// Query for specific metric
		resp, err := server.QueryMetrics(context.Background(), &pb.QueryMetricsRequest{
			Query: "metric1",
		})

		if err != nil {
			t.Fatalf("QueryMetrics failed: %v", err)
		}

		if resp.Status != "success" {
			t.Errorf("Expected status 'success', got '%s'", resp.Status)
		}

		if len(resp.Data) != 1 {
			t.Errorf("Expected 1 metric, got %d", len(resp.Data))
		}

		if resp.Data[0].Name != "metric1" {
			t.Errorf("Expected metric1, got %s", resp.Data[0].Name)
		}
	})

	t.Run("QueryMetrics returns all when query is empty", func(t *testing.T) {
		registry.Clear()

		registry.Register(&metrics.Metric{
			Name:  "metric1",
			Type:  metrics.MetricTypeCounter,
			Value: 10.0,
		})

		registry.Register(&metrics.Metric{
			Name:  "metric2",
			Type:  metrics.MetricTypeGauge,
			Value: 20.0,
		})

		resp, err := server.QueryMetrics(context.Background(), &pb.QueryMetricsRequest{
			Query: "",
		})

		if err != nil {
			t.Fatalf("QueryMetrics failed: %v", err)
		}

		if len(resp.Data) != 2 {
			t.Errorf("Expected 2 metrics, got %d", len(resp.Data))
		}
	})

	t.Run("ListMetrics returns all metrics", func(t *testing.T) {
		registry.Clear()

		registry.Register(&metrics.Metric{
			Name:  "test_metric",
			Type:  metrics.MetricTypeGauge,
			Value: 100.0,
			Labels: map[string]string{
				"env": "test",
			},
		})

		resp, err := server.ListMetrics(context.Background(), &pb.ListMetricsRequest{})

		if err != nil {
			t.Fatalf("ListMetrics failed: %v", err)
		}

		if len(resp.Metrics) != 1 {
			t.Errorf("Expected 1 metric, got %d", len(resp.Metrics))
		}

		metric := resp.Metrics[0]
		if metric.Name != "test_metric" {
			t.Errorf("Expected name 'test_metric', got '%s'", metric.Name)
		}

		if metric.Type != "gauge" {
			t.Errorf("Expected type 'gauge', got '%s'", metric.Type)
		}

		if metric.Value != 100.0 {
			t.Errorf("Expected value 100.0, got %f", metric.Value)
		}

		if metric.Labels["env"] != "test" {
			t.Error("Labels not set correctly")
		}
	})

	t.Run("ListMetrics filters by name", func(t *testing.T) {
		registry.Clear()

		registry.Register(&metrics.Metric{
			Name:  "http_requests",
			Type:  metrics.MetricTypeCounter,
			Value: 100.0,
		})

		registry.Register(&metrics.Metric{
			Name:  "cpu_usage",
			Type:  metrics.MetricTypeGauge,
			Value: 50.0,
		})

		resp, err := server.ListMetrics(context.Background(), &pb.ListMetricsRequest{
			Filter: "http_requests",
		})

		if err != nil {
			t.Fatalf("ListMetrics failed: %v", err)
		}

		if len(resp.Metrics) != 1 {
			t.Errorf("Expected 1 metric, got %d", len(resp.Metrics))
		}

		if resp.Metrics[0].Name != "http_requests" {
			t.Error("Filter not working correctly")
		}
	})

	t.Run("Timestamp is included in responses", func(t *testing.T) {
		registry.Clear()

		before := time.Now().Unix()
		registry.Register(&metrics.Metric{
			Name:  "time_test",
			Type:  metrics.MetricTypeCounter,
			Value: 1.0,
		})
		after := time.Now().Unix()

		resp, err := server.ListMetrics(context.Background(), &pb.ListMetricsRequest{})

		if err != nil {
			t.Fatalf("ListMetrics failed: %v", err)
		}

		if len(resp.Metrics) != 1 {
			t.Fatalf("Expected 1 metric, got %d", len(resp.Metrics))
		}

		timestamp := resp.Metrics[0].Timestamp
		if timestamp < before || timestamp > after {
			t.Error("Timestamp not set correctly")
		}
	})
}
