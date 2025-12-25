package scraper

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Avinash7390/Promenitheus/pkg/config"
	"github.com/Avinash7390/Promenitheus/pkg/metrics"
)

// Scraper handles metric scraping from targets
type Scraper struct {
	config   *config.Config
	registry *metrics.MetricRegistry
	client   *http.Client
}

// NewScraper creates a new scraper
func NewScraper(cfg *config.Config, registry *metrics.MetricRegistry) *Scraper {
	return &Scraper{
		config:   cfg,
		registry: registry,
		client: &http.Client{
			Timeout: cfg.Global.ScrapeTimeout,
		},
	}
}

// Start begins scraping metrics from all configured targets
func (s *Scraper) Start(ctx context.Context) {
	for _, scrapeConfig := range s.config.ScrapeConfigs {
		go s.runScrapeLoop(ctx, scrapeConfig)
	}
}

// runScrapeLoop runs the scrape loop for a specific job
func (s *Scraper) runScrapeLoop(ctx context.Context, cfg config.ScrapeConfig) {
	ticker := time.NewTicker(cfg.ScrapeInterval)
	defer ticker.Stop()

	// Initial scrape
	s.scrapeJob(cfg)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.scrapeJob(cfg)
		}
	}
}

// scrapeJob scrapes all targets in a job
func (s *Scraper) scrapeJob(cfg config.ScrapeConfig) {
	for _, staticConfig := range cfg.StaticConfigs {
		for _, target := range staticConfig.Targets {
			go s.scrapeTarget(target, cfg.JobName, staticConfig.Labels)
		}
	}
}

// scrapeTarget scrapes metrics from a single target
func (s *Scraper) scrapeTarget(target, jobName string, labels map[string]string) {
	url := fmt.Sprintf("http://%s/metrics", target)

	resp, err := s.client.Get(url)
	if err != nil {
		fmt.Printf("Error scraping %s: %v\n", target, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Non-OK status from %s: %d\n", target, resp.StatusCode)
		return
	}

	parsedMetrics, err := s.parseMetrics(resp.Body)
	if err != nil {
		fmt.Printf("Error parsing metrics from %s: %v\n", target, err)
		return
	}

	// Add job and instance labels
	for _, metric := range parsedMetrics {
		if metric.Labels == nil {
			metric.Labels = make(map[string]string)
		}
		metric.Labels["job"] = jobName
		metric.Labels["instance"] = target

		// Add static config labels
		for k, v := range labels {
			metric.Labels[k] = v
		}

		s.registry.Register(metric)
	}
}

// parseMetrics parses Prometheus text format metrics
func (s *Scraper) parseMetrics(r io.Reader) ([]*metrics.Metric, error) {
	var result []*metrics.Metric
	scanner := bufio.NewScanner(r)

	var currentType metrics.MetricType = metrics.MetricTypeGauge

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments (except TYPE)
		if line == "" {
			continue
		}

		// Parse TYPE hints
		if strings.HasPrefix(line, "# TYPE ") {
			parts := strings.Fields(line)
			if len(parts) >= 4 {
				switch parts[3] {
				case "counter":
					currentType = metrics.MetricTypeCounter
				case "gauge":
					currentType = metrics.MetricTypeGauge
				default:
					currentType = metrics.MetricTypeGauge
				}
			}
			continue
		}

		// Skip other comments
		if strings.HasPrefix(line, "#") {
			continue
		}

		// Parse metric line
		metric, err := s.parseMetricLine(line, currentType)
		if err != nil {
			// Skip invalid lines
			continue
		}

		result = append(result, metric)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// parseMetricLine parses a single metric line
func (s *Scraper) parseMetricLine(line string, metricType metrics.MetricType) (*metrics.Metric, error) {
	// Format: metric_name{label1="value1",label2="value2"} value
	// or: metric_name value

	var name string
	var labelsStr string
	var valueStr string

	// Check if line has labels
	if idx := strings.Index(line, "{"); idx != -1 {
		name = strings.TrimSpace(line[:idx])
		rest := line[idx+1:]

		// Find closing brace
		closeIdx := strings.Index(rest, "}")
		if closeIdx == -1 {
			return nil, fmt.Errorf("malformed metric line: missing closing brace")
		}

		labelsStr = rest[:closeIdx]
		valueStr = strings.TrimSpace(rest[closeIdx+1:])
	} else {
		// No labels
		parts := strings.Fields(line)
		if len(parts) < 2 {
			return nil, fmt.Errorf("malformed metric line: %s", line)
		}
		name = parts[0]
		valueStr = parts[1]
	}

	// Parse value
	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse value: %w", err)
	}

	// Parse labels
	labels := make(map[string]string)
	if labelsStr != "" {
		pairs := strings.Split(labelsStr, ",")
		for _, pair := range pairs {
			pair = strings.TrimSpace(pair)
			if pair == "" {
				continue
			}

			kv := strings.SplitN(pair, "=", 2)
			if len(kv) != 2 {
				continue
			}

			key := strings.TrimSpace(kv[0])
			val := strings.Trim(strings.TrimSpace(kv[1]), "\"")
			labels[key] = val
		}
	}

	return &metrics.Metric{
		Name:   name,
		Type:   metricType,
		Value:  value,
		Labels: labels,
	}, nil
}
