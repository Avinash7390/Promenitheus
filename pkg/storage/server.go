package storage

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/soheilhy/cmux"
	pb "github.com/Avinash7390/Promenitheus/api/proto/v1"
	"github.com/Avinash7390/Promenitheus/pkg/grpcserver"
	"github.com/Avinash7390/Promenitheus/pkg/metrics"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Server exposes stored metrics via HTTP and gRPC on the same port using cmux
type Server struct {
	registry   *metrics.MetricRegistry
	port       int
	grpcServer *grpc.Server
	httpServer *http.Server
	listener   net.Listener
	mux        cmux.CMux
}

// NewServer creates a new storage server
func NewServer(registry *metrics.MetricRegistry, port int) *Server {
	return &Server{
		registry: registry,
		port:     port,
	}
}

// Start starts both HTTP and gRPC servers on the same port using cmux
func (s *Server) Start() error {
	// Create a TCP listener
	var err error
	s.listener, err = net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", s.port, err)
	}

	// Create a cmux instance
	s.mux = cmux.New(s.listener)

	// Match HTTP/2 connections for gRPC
	grpcL := s.mux.MatchWithWriters(cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"))

	// Match HTTP/1.x connections
	httpL := s.mux.Match(cmux.HTTP1Fast())

	// Setup gRPC server
	s.grpcServer = grpc.NewServer()
	metricsServer := grpcserver.NewMetricsServer(s.registry)
	pb.RegisterMetricsServiceServer(s.grpcServer, metricsServer)
	reflection.Register(s.grpcServer)

	// Setup gRPC-Gateway (HTTP to gRPC translator)
	// Use in-process connection instead of dialing back to ourselves
	gwmux := runtime.NewServeMux()
	
	// Register gRPC-Gateway handlers with the server directly (no dial needed)
	err = pb.RegisterMetricsServiceHandlerServer(context.Background(), gwmux, metricsServer)
	if err != nil {
		return fmt.Errorf("failed to register gateway: %w", err)
	}

	// Setup HTTP server with custom handlers and gRPC-Gateway fallback
	httpMux := http.NewServeMux()
	
	// Custom /metrics handler for Prometheus text format
	httpMux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		// Call gRPC method directly
		resp, err := metricsServer.GetMetrics(r.Context(), &pb.GetMetricsRequest{})
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", resp.ContentType)
		w.Write([]byte(resp.Content))
	})
	
	// Custom homepage handler
	httpMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			s.handleIndex(w, r)
			return
		}
		// Forward other requests to gRPC-Gateway
		gwmux.ServeHTTP(w, r)
	})

	s.httpServer = &http.Server{
		Handler: httpMux,
	}

	fmt.Printf("Starting unified server on :%d (HTTP/1.1 and gRPC/HTTP2)\n", s.port)
	fmt.Printf("  - HTTP/1.1 requests → HTTP handlers\n")
	fmt.Printf("  - HTTP/2 gRPC requests → gRPC server\n")
	fmt.Printf("  - HTTP/JSON requests → gRPC via grpc-gateway\n")

	// Start serving gRPC and HTTP
	errChan := make(chan error, 3)

	// Serve gRPC
	go func() {
		if err := s.grpcServer.Serve(grpcL); err != nil {
			errChan <- fmt.Errorf("gRPC server error: %w", err)
		}
	}()

	// Serve HTTP
	go func() {
		if err := s.httpServer.Serve(httpL); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("HTTP server error: %w", err)
		}
	}()

	// Start cmux
	go func() {
		if err := s.mux.Serve(); err != nil {
			errChan <- fmt.Errorf("cmux error: %w", err)
		}
	}()

	// Wait for any error
	return <-errChan
}

// Shutdown gracefully stops both servers
func (s *Server) Shutdown(ctx context.Context) error {
	// Stop gRPC server
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}

	// Stop HTTP server
	if s.httpServer != nil {
		if err := s.httpServer.Shutdown(ctx); err != nil {
			return err
		}
	}

	// Close listener
	if s.listener != nil {
		return s.listener.Close()
	}

	return nil
}

// handleIndex serves a simple home page
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Promenitheus</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        h1 { color: #333; }
        h2 { color: #666; margin-top: 30px; }
        .highlight { background-color: #ffeb3b; padding: 2px 5px; }
        pre { background-color: #f5f5f5; padding: 10px; border-radius: 4px; }
        .info { background-color: #e3f2fd; padding: 15px; border-radius: 4px; margin: 20px 0; }
    </style>
</head>
<body>
    <h1>Promenitheus - Prometheus-like Metric Scraper</h1>
    
    <div class="info">
        <strong>🚀 Single Port Architecture</strong><br>
        All APIs accessible on <span class="highlight">port ` + fmt.Sprintf("%d", s.port) + `</span> using connection multiplexing (cmux)<br>
        • HTTP/1.1 requests → HTTP handlers<br>
        • HTTP/2 gRPC requests → gRPC server<br>
        • HTTP/JSON requests → gRPC via grpc-gateway
    </div>

    <h2>HTTP/REST API (HTTP/1.1)</h2>
    <ul>
        <li><a href="/metrics">GET /metrics</a> - Metrics in Prometheus text format</li>
        <li><a href="/api/v1/query">GET /api/v1/query?query=metric_name</a> - Query specific metrics (JSON)</li>
        <li><a href="/api/v1/metrics">GET /api/v1/metrics?filter=metric_name</a> - List all metrics (JSON)</li>
    </ul>

    <h2>gRPC API (HTTP/2)</h2>
    <p>Same port, different protocol. Use gRPC clients or grpcurl:</p>
    <ul>
        <li><code>MetricsService.GetMetrics</code> - Get metrics in Prometheus format</li>
        <li><code>MetricsService.QueryMetrics</code> - Query metrics by name</li>
        <li><code>MetricsService.ListMetrics</code> - List all metrics</li>
    </ul>

    <h3>Testing with grpcurl</h3>
    <pre>grpcurl -plaintext localhost:` + fmt.Sprintf("%d", s.port) + ` list
grpcurl -plaintext localhost:` + fmt.Sprintf("%d", s.port) + ` promenitheus.v1.MetricsService/GetMetrics
grpcurl -plaintext -d '{"query": "http_requests_total"}' localhost:` + fmt.Sprintf("%d", s.port) + ` promenitheus.v1.MetricsService/QueryMetrics</pre>

    <h3>Testing with curl (HTTP/JSON → gRPC via gateway)</h3>
    <pre>curl http://localhost:` + fmt.Sprintf("%d", s.port) + `/metrics
curl http://localhost:` + fmt.Sprintf("%d", s.port) + `/api/v1/query?query=http_requests_total
curl http://localhost:` + fmt.Sprintf("%d", s.port) + `/api/v1/metrics</pre>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}
