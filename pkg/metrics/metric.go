package metrics

import (
	"fmt"
	"sync"
	"time"
)

// MetricType represents the type of metric
type MetricType string

const (
	MetricTypeCounter MetricType = "counter"
	MetricTypeGauge   MetricType = "gauge"
)

// Metric represents a single metric with labels
type Metric struct {
	Name      string            `json:"name"`
	Type      MetricType        `json:"type"`
	Value     float64           `json:"value"`
	Labels    map[string]string `json:"labels,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
}

// MetricRegistry stores and manages metrics
type MetricRegistry struct {
	mu      sync.RWMutex
	metrics map[string]*Metric
}

// NewMetricRegistry creates a new metric registry
func NewMetricRegistry() *MetricRegistry {
	return &MetricRegistry{
		metrics: make(map[string]*Metric),
	}
}

// Register adds or updates a metric in the registry
func (r *MetricRegistry) Register(metric *Metric) {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := r.generateKey(metric.Name, metric.Labels)
	metric.Timestamp = time.Now()
	r.metrics[key] = metric
}

// Get retrieves a metric by name and labels
func (r *MetricRegistry) Get(name string, labels map[string]string) (*Metric, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key := r.generateKey(name, labels)
	metric, exists := r.metrics[key]
	return metric, exists
}

// GetAll returns all metrics in the registry
func (r *MetricRegistry) GetAll() []*Metric {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*Metric, 0, len(r.metrics))
	for _, metric := range r.metrics {
		result = append(result, metric)
	}
	return result
}

// Clear removes all metrics from the registry
func (r *MetricRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.metrics = make(map[string]*Metric)
}

// generateKey creates a unique key for a metric based on name and labels
func (r *MetricRegistry) generateKey(name string, labels map[string]string) string {
	if len(labels) == 0 {
		return name
	}

	key := name
	for k, v := range labels {
		key += fmt.Sprintf(",%s=%s", k, v)
	}
	return key
}
