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

	// Setup HTTP server - ALL requests go through gRPC-Gateway
	// No custom HTTP handlers - everything is routed to gRPC handlers
	s.httpServer = &http.Server{
		Handler: gwmux,
	}

	fmt.Printf("Starting unified server on :%d (HTTP/1.1 and gRPC/HTTP2)\n", s.port)
	fmt.Printf("  - HTTP/1.1 requests → grpc-gateway → gRPC handlers\n")
	fmt.Printf("  - HTTP/2 gRPC requests → gRPC server (direct)\n")
	fmt.Printf("  - All HTTP routes handled by grpc-gateway (no custom HTTP handlers)\n")

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
