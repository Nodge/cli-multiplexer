package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadFromFile loads configuration from a file (JSON or YAML)
func LoadFromFile(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file '%s': %w", path, err)
	}
	defer file.Close()

	return LoadFromReader(file, filepath.Ext(path))
}

// LoadFromStdin loads configuration from stdin
func LoadFromStdin(format string) (*Config, error) {
	if format == "" {
		format = "json"
	}
	return LoadFromReader(os.Stdin, format)
}

// LoadFromReader loads configuration from an io.Reader
func LoadFromReader(reader io.Reader, format string) (*Config, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read config data: %w", err)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("config data is empty")
	}

	// Default to JSON if format is empty
	if format == "" {
		format = "json"
	}

	var config Config

	switch strings.ToLower(format) {
	case ".json", "json":
		if err := json.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("failed to parse JSON config: %w", err)
		}
	case ".yaml", ".yml", "yaml", "yml":
		if err := yaml.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("failed to parse YAML config: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported config format '%s', supported formats: json, yaml, yml", format)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// Load is a convenience function that loads config from file or stdin
func Load(configPath string, fromStdin bool, format string) (*Config, error) {
	if fromStdin {
		return LoadFromStdin(format)
	}

	if configPath == "" {
		return nil, fmt.Errorf("config path cannot be empty")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file '%s' does not exist", configPath)
	}

	return LoadFromFile(configPath)
}
