package config

import (
	"os"

	"github.com/stillunbuilt/mutefeed/internal/filter"
	"gopkg.in/yaml.v3"
)

// NamedFeed allows defining static named feed URLs with custom options
type NamedFeed struct {
	URL     string         `json:"url" yaml:"url"`
	Options filter.Options `json:"options" yaml:"options"`
}

// Config represents application configuration
type Config struct {
	Host        string               `json:"host" yaml:"host"`
	Port        int                  `json:"port" yaml:"port"`
	Defaults    filter.Options       `json:"defaults" yaml:"defaults"`
	Feeds       map[string]NamedFeed `json:"feeds" yaml:"feeds"`
	MetricsFile string               `json:"metrics_file" yaml:"metrics_file"`
}

// DefaultConfig returns clean default settings
func DefaultConfig() *Config {
	return &Config{
		Host:     "0.0.0.0",
		Port:     8899,
		Defaults: filter.DefaultOptions(),
		Feeds: map[string]NamedFeed{
			"hn": {
				URL: "https://news.ycombinator.com/rss",
				Options: filter.Options{
					Presets:       []string{"paywalls", "ai-slop", "clickbait", "sponsored"},
					FilterSlop:    true,
					SlopThreshold: 0.5,
					CleanLinks:    true,
					MaxItems:      50,
				},
			},
		},
	}
}

// LoadConfig loads configuration from path or returns default if not found
func LoadConfig(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
