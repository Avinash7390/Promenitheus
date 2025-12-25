package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the main configuration
type Config struct {
	Global        GlobalConfig   `yaml:"global"`
	ScrapeConfigs []ScrapeConfig `yaml:"scrape_configs"`
}

// GlobalConfig contains global settings
type GlobalConfig struct {
	ScrapeInterval time.Duration `yaml:"scrape_interval"`
	ScrapeTimeout  time.Duration `yaml:"scrape_timeout"`
}

// ScrapeConfig defines a scrape job
type ScrapeConfig struct {
	JobName       string          `yaml:"job_name"`
	ScrapeInterval time.Duration  `yaml:"scrape_interval,omitempty"`
	ScrapeTimeout  time.Duration  `yaml:"scrape_timeout,omitempty"`
	StaticConfigs []StaticConfig  `yaml:"static_configs"`
}

// StaticConfig defines static targets
type StaticConfig struct {
	Targets []string          `yaml:"targets"`
	Labels  map[string]string `yaml:"labels,omitempty"`
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults
	if config.Global.ScrapeInterval == 0 {
		config.Global.ScrapeInterval = 15 * time.Second
	}
	if config.Global.ScrapeTimeout == 0 {
		config.Global.ScrapeTimeout = 10 * time.Second
	}

	// Apply global defaults to scrape configs
	for i := range config.ScrapeConfigs {
		if config.ScrapeConfigs[i].ScrapeInterval == 0 {
			config.ScrapeConfigs[i].ScrapeInterval = config.Global.ScrapeInterval
		}
		if config.ScrapeConfigs[i].ScrapeTimeout == 0 {
			config.ScrapeConfigs[i].ScrapeTimeout = config.Global.ScrapeTimeout
		}
	}

	return &config, nil
}
