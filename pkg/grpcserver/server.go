package grpcserver

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strings"

	pb "github.com/Avinash7390/Promenitheus/api/proto/v1"
	"github.com/Avinash7390/Promenitheus/pkg/metrics"
)

type MetricsServer struct {
	pb.UnimplementedMetricsServiceServer
	registry *metrics.MetricRegistry
}

func NewMetricsServer(registry *metrics.MetricRegistry) *MetricsServer {
	return &MetricsServer{
		registry: registry,
	}
}

func (s *MetricsServer) GetMetrics(ctx context.Context, req *pb.GetMetricsRequest) (*pb.GetMetricsResponse, error) {
	allMetrics := s.registry.GetAll()

	// Sort metrics by name for consistent output
	sort.Slice(allMetrics, func(i, j int) bool {
		return allMetrics[i].Name < allMetrics[j].Name
	})

	var buf bytes.Buffer

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
			fmt.Fprintf(&buf, "# TYPE %s %s\n", m.Name, m.Type)
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

			fmt.Fprintf(&buf, "%s{%s} %v\n", m.Name, strings.Join(labelPairs, ","), m.Value)
		} else {
			fmt.Fprintf(&buf, "%s %v\n", m.Name, m.Value)
		}
	}

	return &pb.GetMetricsResponse{
		Content:     buf.String(),
		ContentType: "text/plain; version=0.0.4",
	}, nil
}

// QueryMetrics queries metrics by name
func (s *MetricsServer) QueryMetrics(ctx context.Context, req *pb.QueryMetricsRequest) (*pb.QueryMetricsResponse, error) {
	var result []*pb.Metric

	allMetrics := s.registry.GetAll()
	for _, m := range allMetrics {
		if req.Query == "" || m.Name == req.Query {
			result = append(result, &pb.Metric{
				Name:      m.Name,
				Type:      string(m.Type),
				Value:     m.Value,
				Labels:    m.Labels,
				Timestamp: m.Timestamp.Unix(),
			})
		}
	}

	return &pb.QueryMetricsResponse{
		Status: "success",
		Data:   result,
	}, nil
}

func (s *MetricsServer) ListMetrics(ctx context.Context, req *pb.ListMetricsRequest) (*pb.ListMetricsResponse, error) {
	var result []*pb.Metric

	allMetrics := s.registry.GetAll()
	for _, m := range allMetrics {
		if req.Filter == "" || m.Name == req.Filter {
			result = append(result, &pb.Metric{
				Name:      m.Name,
				Type:      string(m.Type),
				Value:     m.Value,
				Labels:    m.Labels,
				Timestamp: m.Timestamp.Unix(),
			})
		}
	}

	return &pb.ListMetricsResponse{
		Metrics: result,
	}, nil
}
