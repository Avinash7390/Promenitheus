package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sort"
	"strings"

	pb "github.com/Avinash7390/Promenitheus/api/proto/v1"
	"github.com/Avinash7390/Promenitheus/pkg/grpcserver"
	"github.com/Avinash7390/Promenitheus/pkg/metrics"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Server exposes stored metrics via HTTP and gRPC
type Server struct {
	registry   *metrics.MetricRegistry
	port       int
	grpcPort   int
	httpServer *http.Server
	grpcServer *grpc.Server
}

// NewServer creates a new storage server
func NewServer(registry *metrics.MetricRegistry, port int) *Server {
	return &Server{
		registry: registry,
		port:     port,
		grpcPort: port + 1, // gRPC on port+1 by default
	}
}

// Start starts both HTTP and gRPC servers
func (s *Server) Start() error {
	// Start gRPC server in a goroutine
	errChan := make(chan error, 2)
	
	go func() {
		if err := s.startGRPC(); err != nil {
			errChan <- fmt.Errorf("gRPC server error: %w", err)
		}
	}()

	// Start HTTP server
	go func() {
		if err := s.startHTTP(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("HTTP server error: %w", err)
		}
	}()

	// Wait for any error
	return <-errChan
}

// startGRPC starts the gRPC server
func (s *Server) startGRPC() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.grpcPort))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", s.grpcPort, err)
	}

	s.grpcServer = grpc.NewServer()
	
	// Register the metrics service
	metricsServer := grpcserver.NewMetricsServer(s.registry)
	pb.RegisterMetricsServiceServer(s.grpcServer, metricsServer)
	
	// Register reflection service for grpcurl/grpc_cli
	reflection.Register(s.grpcServer)

	fmt.Printf("Starting gRPC server on :%d\n", s.grpcPort)
	return s.grpcServer.Serve(lis)
}

// startHTTP starts the HTTP server
func (s *Server) startHTTP() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", s.handleMetrics)
	mux.HandleFunc("/api/v1/query", s.handleQuery)
	mux.HandleFunc("/api/v1/metrics", s.handleListMetrics)
	mux.HandleFunc("/", s.handleIndex)

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: mux,
	}

	fmt.Printf("Starting HTTP server on :%d\n", s.port)
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully stops both servers
func (s *Server) Shutdown(ctx context.Context) error {
	// Stop gRPC server
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}

	// Stop HTTP server
	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}

	return nil
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
    <h2>HTTP API (Port ` + fmt.Sprintf("%d", s.port) + `)</h2>
    <ul>
        <li><a href="/metrics">GET /metrics - Metrics (Prometheus format)</a></li>
        <li><a href="/api/v1/query">GET /api/v1/query - Query API (JSON)</a></li>
        <li><a href="/api/v1/metrics">GET /api/v1/metrics - List all metrics (JSON)</a></li>
    </ul>
    <h2>gRPC API (Port ` + fmt.Sprintf("%d", s.grpcPort) + `)</h2>
    <ul>
        <li>MetricsService.GetMetrics - Get metrics in Prometheus format</li>
        <li>MetricsService.QueryMetrics - Query metrics by name</li>
        <li>MetricsService.ListMetrics - List all metrics</li>
    </ul>
    <p>Use <code>grpcurl</code> to test gRPC endpoints:</p>
    <pre>grpcurl -plaintext localhost:` + fmt.Sprintf("%d", s.grpcPort) + ` list
grpcurl -plaintext localhost:` + fmt.Sprintf("%d", s.grpcPort) + ` promenitheus.v1.MetricsService/GetMetrics</pre>
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

// handleListMetrics returns all metrics in JSON format
func (s *Server) handleListMetrics(w http.ResponseWriter, r *http.Request) {
	filter := r.URL.Query().Get("filter")

	var result []*metrics.Metric
	allMetrics := s.registry.GetAll()

	for _, m := range allMetrics {
		if filter == "" || m.Name == filter {
			result = append(result, m)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"metrics": result,
	})
}
