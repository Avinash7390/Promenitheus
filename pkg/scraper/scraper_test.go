package scraper

import (
	"strings"
	"testing"

	"github.com/Avinash7390/Promenitheus/pkg/metrics"
)

func TestParseMetrics(t *testing.T) {
	scraper := &Scraper{}

	t.Run("Parse simple metric without labels", func(t *testing.T) {
		input := `# TYPE simple_metric counter
simple_metric 42`

		reader := strings.NewReader(input)
		parsed, err := scraper.parseMetrics(reader)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if len(parsed) != 1 {
			t.Fatalf("Expected 1 metric, got %d", len(parsed))
		}

		if parsed[0].Name != "simple_metric" {
			t.Errorf("Expected name 'simple_metric', got '%s'", parsed[0].Name)
		}

		if parsed[0].Value != 42 {
			t.Errorf("Expected value 42, got %f", parsed[0].Value)
		}

		if parsed[0].Type != metrics.MetricTypeCounter {
			t.Errorf("Expected type counter, got %s", parsed[0].Type)
		}
	})

	t.Run("Parse metric with labels", func(t *testing.T) {
		input := `# TYPE http_requests counter
http_requests{method="GET",path="/api"} 100`

		reader := strings.NewReader(input)
		parsed, err := scraper.parseMetrics(reader)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if len(parsed) != 1 {
			t.Fatalf("Expected 1 metric, got %d", len(parsed))
		}

		metric := parsed[0]
		if metric.Name != "http_requests" {
			t.Errorf("Expected name 'http_requests', got '%s'", metric.Name)
		}

		if metric.Value != 100 {
			t.Errorf("Expected value 100, got %f", metric.Value)
		}

		if metric.Labels["method"] != "GET" {
			t.Errorf("Expected label method=GET, got %s", metric.Labels["method"])
		}

		if metric.Labels["path"] != "/api" {
			t.Errorf("Expected label path=/api, got %s", metric.Labels["path"])
		}
	})

	t.Run("Parse multiple metrics", func(t *testing.T) {
		input := `# TYPE metric1 counter
metric1 10
# TYPE metric2 gauge
metric2 20.5
# TYPE metric3 counter
metric3{label="value"} 30`

		reader := strings.NewReader(input)
		parsed, err := scraper.parseMetrics(reader)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if len(parsed) != 3 {
			t.Fatalf("Expected 3 metrics, got %d", len(parsed))
		}

		if parsed[0].Name != "metric1" || parsed[0].Value != 10 {
			t.Error("First metric incorrect")
		}

		if parsed[1].Name != "metric2" || parsed[1].Value != 20.5 {
			t.Error("Second metric incorrect")
		}

		if parsed[2].Name != "metric3" || parsed[2].Value != 30 {
			t.Error("Third metric incorrect")
		}
	})

	t.Run("Skip comments and empty lines", func(t *testing.T) {
		input := `# This is a comment
# TYPE test_metric counter

test_metric 123

# Another comment`

		reader := strings.NewReader(input)
		parsed, err := scraper.parseMetrics(reader)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if len(parsed) != 1 {
			t.Fatalf("Expected 1 metric, got %d", len(parsed))
		}

		if parsed[0].Name != "test_metric" || parsed[0].Value != 123 {
			t.Error("Metric not parsed correctly")
		}
	})

	t.Run("Handle gauge type", func(t *testing.T) {
		input := `# TYPE gauge_metric gauge
gauge_metric 3.14`

		reader := strings.NewReader(input)
		parsed, err := scraper.parseMetrics(reader)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if parsed[0].Type != metrics.MetricTypeGauge {
			t.Errorf("Expected gauge type, got %s", parsed[0].Type)
		}
	})

	t.Run("Parse metric with multiple labels", func(t *testing.T) {
		input := `http_requests{method="POST",status="200",path="/users"} 456`

		reader := strings.NewReader(input)
		parsed, err := scraper.parseMetrics(reader)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if len(parsed) != 1 {
			t.Fatalf("Expected 1 metric, got %d", len(parsed))
		}

		metric := parsed[0]
		if len(metric.Labels) != 3 {
			t.Errorf("Expected 3 labels, got %d", len(metric.Labels))
		}

		if metric.Labels["method"] != "POST" {
			t.Error("method label incorrect")
		}
		if metric.Labels["status"] != "200" {
			t.Error("status label incorrect")
		}
		if metric.Labels["path"] != "/users" {
			t.Error("path label incorrect")
		}
	})
}
