package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

// Config holds all configuration settings for the application
type Config struct {
	Server struct {
		Port int    `json:"port" env:"SERVER_PORT"`
		Host string `json:"host" env:"SERVER_HOST"`
	} `json:"server"`

	ACME struct {
		DirectoryURL string `json:"directoryURL" env:"ACME_DIRECTORY_URL"`
		Environment  string `json:"environment" env:"ACME_ENVIRONMENT"`
	} `json:"acme"`
}

// DefaultConfig returns a configuration with default values
func DefaultConfig() *Config {
	cfg := &Config{}
	cfg.Server.Port = 8080
	cfg.Server.Host = "localhost"
	cfg.ACME.Environment = "development"
	cfg.ACME.DirectoryURL = "https://acme-staging-v02.api.letsencrypt.org/directory"
	return cfg
}

// Load reads configuration from a JSON file and environment variables
func Load(filepath string) (*Config, error) {
	// Start with default configuration
	cfg := DefaultConfig()

	// If filepath is provided, load configuration from JSON file
	if filepath != "" {
		file, err := os.ReadFile(filepath)
		if err != nil {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}

		if err := json.Unmarshal(file, cfg); err != nil {
			return nil, fmt.Errorf("error parsing config file: %w", err)
		}
	}

	// Override with environment variables
	if err := loadEnvVars(cfg); err != nil {
		return nil, fmt.Errorf("error loading environment variables: %w", err)
	}

	// Validate the configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// loadEnvVars overrides configuration with environment variables
func loadEnvVars(cfg *Config) error {
	// Server configuration
	if port := os.Getenv("SERVER_PORT"); port != "" {
		p, err := strconv.Atoi(port)
		if err != nil {
			return fmt.Errorf("invalid SERVER_PORT: %w", err)
		}
		cfg.Server.Port = p
	}

	if host := os.Getenv("SERVER_HOST"); host != "" {
		cfg.Server.Host = host
	}

	// ACME configuration
	if url := os.Getenv("ACME_DIRECTORY_URL"); url != "" {
		cfg.ACME.DirectoryURL = url
	}

	if env := os.Getenv("ACME_ENVIRONMENT"); env != "" {
		cfg.ACME.Environment = env
	}

	return nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid port number: %d", c.Server.Port)
	}

	if c.Server.Host == "" {
		return fmt.Errorf("host cannot be empty")
	}

	if c.ACME.Environment != "development" && c.ACME.Environment != "production" {
		return fmt.Errorf("invalid environment: %s", c.ACME.Environment)
	}

	return nil
}

// String returns a string representation of the configuration
func (c *Config) String() string {
	b, _ := json.MarshalIndent(c, "", "  ")
	return string(b)
}
