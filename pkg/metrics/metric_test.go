package metrics

import (
	"testing"
	"time"
)

func TestMetricRegistry(t *testing.T) {
	registry := NewMetricRegistry()

	t.Run("Register and Get metric", func(t *testing.T) {
		metric := &Metric{
			Name:  "test_metric",
			Type:  MetricTypeCounter,
			Value: 42.0,
			Labels: map[string]string{
				"label1": "value1",
			},
		}

		registry.Register(metric)

		retrieved, exists := registry.Get("test_metric", map[string]string{"label1": "value1"})
		if !exists {
			t.Fatal("Expected metric to exist")
		}

		if retrieved.Name != "test_metric" {
			t.Errorf("Expected name 'test_metric', got '%s'", retrieved.Name)
		}

		if retrieved.Value != 42.0 {
			t.Errorf("Expected value 42.0, got %f", retrieved.Value)
		}
	})

	t.Run("GetAll metrics", func(t *testing.T) {
		registry.Clear()

		metrics := []*Metric{
			{Name: "metric1", Type: MetricTypeCounter, Value: 1.0},
			{Name: "metric2", Type: MetricTypeGauge, Value: 2.0},
			{Name: "metric3", Type: MetricTypeCounter, Value: 3.0},
		}

		for _, m := range metrics {
			registry.Register(m)
		}

		all := registry.GetAll()
		if len(all) != 3 {
			t.Errorf("Expected 3 metrics, got %d", len(all))
		}
	})

	t.Run("Update existing metric", func(t *testing.T) {
		registry.Clear()

		metric1 := &Metric{
			Name:  "update_test",
			Type:  MetricTypeCounter,
			Value: 10.0,
		}
		registry.Register(metric1)

		metric2 := &Metric{
			Name:  "update_test",
			Type:  MetricTypeCounter,
			Value: 20.0,
		}
		registry.Register(metric2)

		retrieved, exists := registry.Get("update_test", nil)
		if !exists {
			t.Fatal("Expected metric to exist")
		}

		if retrieved.Value != 20.0 {
			t.Errorf("Expected updated value 20.0, got %f", retrieved.Value)
		}
	})

	t.Run("Metrics with different labels", func(t *testing.T) {
		registry.Clear()

		metric1 := &Metric{
			Name:   "http_requests",
			Type:   MetricTypeCounter,
			Value:  100.0,
			Labels: map[string]string{"method": "GET"},
		}

		metric2 := &Metric{
			Name:   "http_requests",
			Type:   MetricTypeCounter,
			Value:  50.0,
			Labels: map[string]string{"method": "POST"},
		}

		registry.Register(metric1)
		registry.Register(metric2)

		all := registry.GetAll()
		if len(all) != 2 {
			t.Errorf("Expected 2 metrics, got %d", len(all))
		}

		get1, exists := registry.Get("http_requests", map[string]string{"method": "GET"})
		if !exists || get1.Value != 100.0 {
			t.Error("GET metric not found or incorrect value")
		}

		get2, exists := registry.Get("http_requests", map[string]string{"method": "POST"})
		if !exists || get2.Value != 50.0 {
			t.Error("POST metric not found or incorrect value")
		}
	})

	t.Run("Timestamp is set on registration", func(t *testing.T) {
		registry.Clear()

		before := time.Now()
		metric := &Metric{
			Name:  "timestamp_test",
			Type:  MetricTypeGauge,
			Value: 123.0,
		}
		registry.Register(metric)
		after := time.Now()

		retrieved, exists := registry.Get("timestamp_test", nil)
		if !exists {
			t.Fatal("Expected metric to exist")
		}

		if retrieved.Timestamp.Before(before) || retrieved.Timestamp.After(after) {
			t.Error("Timestamp not set correctly")
		}
	})
}
