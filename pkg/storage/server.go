package storage

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/Avinash7390/Promenitheus/pkg/metrics"
)

// Server exposes stored metrics via HTTP
type Server struct {
	registry *metrics.MetricRegistry
	port     int
}

// NewServer creates a new storage server
func NewServer(registry *metrics.MetricRegistry, port int) *Server {
	return &Server{
		registry: registry,
		port:     port,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	http.HandleFunc("/metrics", s.handleMetrics)
	http.HandleFunc("/api/v1/query", s.handleQuery)
	http.HandleFunc("/", s.handleIndex)

	addr := fmt.Sprintf(":%d", s.port)
	fmt.Printf("Starting Promenitheus server on %s\n", addr)
	return http.ListenAndServe(addr, nil)
}

// handleIndex serves a simple home page
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	html := `<!DOCTYPE html>
<html>
<head>
    <title>Promenitheus</title>
</head>
<body>
    <h1>Promenitheus - Prometheus-like Metric Scraper</h1>
    <ul>
        <li><a href="/metrics">Metrics (Prometheus format)</a></li>
        <li><a href="/api/v1/query">Query API</a></li>
    </ul>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// handleMetrics exposes metrics in Prometheus text format
func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	allMetrics := s.registry.GetAll()

	// Sort metrics by name for consistent output
	sort.Slice(allMetrics, func(i, j int) bool {
		return allMetrics[i].Name < allMetrics[j].Name
	})

	w.Header().Set("Content-Type", "text/plain; version=0.0.4")

	// Group metrics by name and type for TYPE hints
	typeMap := make(map[string]metrics.MetricType)
	for _, m := range allMetrics {
		if _, exists := typeMap[m.Name]; !exists {
			typeMap[m.Name] = m.Type
		}
	}

	currentName := ""
	for _, m := range allMetrics {
		// Write TYPE hint when we encounter a new metric name
		if m.Name != currentName {
			fmt.Fprintf(w, "# TYPE %s %s\n", m.Name, m.Type)
			currentName = m.Name
		}

		// Write metric line
		if len(m.Labels) > 0 {
			// Sort labels for consistent output
			labelPairs := make([]string, 0, len(m.Labels))
			for k, v := range m.Labels {
				labelPairs = append(labelPairs, fmt.Sprintf(`%s="%s"`, k, v))
			}
			sort.Strings(labelPairs)

			fmt.Fprintf(w, "%s{%s} %v\n", m.Name, strings.Join(labelPairs, ","), m.Value)
		} else {
			fmt.Fprintf(w, "%s %v\n", m.Name, m.Value)
		}
	}
}

// handleQuery provides a simple query API
func (s *Server) handleQuery(w http.ResponseWriter, r *http.Request) {
	metricName := r.URL.Query().Get("query")

	var result []*metrics.Metric
	if metricName == "" {
		// Return all metrics
		result = s.registry.GetAll()
	} else {
		// Filter by metric name
		allMetrics := s.registry.GetAll()
		for _, m := range allMetrics {
			if m.Name == metricName {
				result = append(result, m)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   result,
	})
}
