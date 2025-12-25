package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	t.Run("Load valid config", func(t *testing.T) {
		configContent := `global:
  scrape_interval: 30s
  scrape_timeout: 15s

scrape_configs:
  - job_name: 'test-job'
    scrape_interval: 10s
    static_configs:
      - targets:
          - 'localhost:8080'
        labels:
          env: 'test'
`

		tmpFile, err := os.CreateTemp("", "config-*.yaml")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tmpFile.Name())

		if _, err := tmpFile.Write([]byte(configContent)); err != nil {
			t.Fatal(err)
		}
		tmpFile.Close()

		cfg, err := LoadConfig(tmpFile.Name())
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if cfg.Global.ScrapeInterval != 30*time.Second {
			t.Errorf("Expected global scrape_interval 30s, got %v", cfg.Global.ScrapeInterval)
		}

		if cfg.Global.ScrapeTimeout != 15*time.Second {
			t.Errorf("Expected global scrape_timeout 15s, got %v", cfg.Global.ScrapeTimeout)
		}

		if len(cfg.ScrapeConfigs) != 1 {
			t.Fatalf("Expected 1 scrape config, got %d", len(cfg.ScrapeConfigs))
		}

		scrapeConfig := cfg.ScrapeConfigs[0]
		if scrapeConfig.JobName != "test-job" {
			t.Errorf("Expected job_name 'test-job', got '%s'", scrapeConfig.JobName)
		}

		if scrapeConfig.ScrapeInterval != 10*time.Second {
			t.Errorf("Expected scrape_interval 10s, got %v", scrapeConfig.ScrapeInterval)
		}

		if len(scrapeConfig.StaticConfigs) != 1 {
			t.Fatalf("Expected 1 static config, got %d", len(scrapeConfig.StaticConfigs))
		}

		staticConfig := scrapeConfig.StaticConfigs[0]
		if len(staticConfig.Targets) != 1 || staticConfig.Targets[0] != "localhost:8080" {
			t.Error("Targets not parsed correctly")
		}

		if staticConfig.Labels["env"] != "test" {
			t.Error("Labels not parsed correctly")
		}
	})

	t.Run("Apply default values", func(t *testing.T) {
		configContent := `scrape_configs:
  - job_name: 'minimal-job'
    static_configs:
      - targets:
          - 'localhost:9000'
`

		tmpFile, err := os.CreateTemp("", "config-*.yaml")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tmpFile.Name())

		if _, err := tmpFile.Write([]byte(configContent)); err != nil {
			t.Fatal(err)
		}
		tmpFile.Close()

		cfg, err := LoadConfig(tmpFile.Name())
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Check defaults are applied
		if cfg.Global.ScrapeInterval != 15*time.Second {
			t.Errorf("Expected default global scrape_interval 15s, got %v", cfg.Global.ScrapeInterval)
		}

		if cfg.Global.ScrapeTimeout != 10*time.Second {
			t.Errorf("Expected default global scrape_timeout 10s, got %v", cfg.Global.ScrapeTimeout)
		}

		// Check that scrape config inherits global defaults
		scrapeConfig := cfg.ScrapeConfigs[0]
		if scrapeConfig.ScrapeInterval != 15*time.Second {
			t.Errorf("Expected inherited scrape_interval 15s, got %v", scrapeConfig.ScrapeInterval)
		}

		if scrapeConfig.ScrapeTimeout != 10*time.Second {
			t.Errorf("Expected inherited scrape_timeout 10s, got %v", scrapeConfig.ScrapeTimeout)
		}
	})

	t.Run("Invalid file path", func(t *testing.T) {
		_, err := LoadConfig("/nonexistent/config.yaml")
		if err == nil {
			t.Error("Expected error for nonexistent file")
		}
	})

	t.Run("Invalid YAML", func(t *testing.T) {
		configContent := `this is not: valid: yaml: content`

		tmpFile, err := os.CreateTemp("", "config-*.yaml")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tmpFile.Name())

		if _, err := tmpFile.Write([]byte(configContent)); err != nil {
			t.Fatal(err)
		}
		tmpFile.Close()

		_, err = LoadConfig(tmpFile.Name())
		if err == nil {
			t.Error("Expected error for invalid YAML")
		}
	})
}
