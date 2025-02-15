package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config holds all configuration settings for the application
type Config struct {
	Server struct {
		ListenAddr string `json:"listen_addr"`
	} `json:"server"`

	ACME struct {
		DirectoryURL string `json:"directoryURL" env:"ACME_DIRECTORY_URL"`
		Environment  string `json:"environment" env:"ACME_ENVIRONMENT"`
	} `json:"acme"`

	CA struct {
		Certs      string `json:"certificates"`
		PrivateKey string `json:"private_key"`
	} `json:"ca"`

	TLS struct {
		Certs      string   `json:"certificates"`
		PrivateKey string   `json:"private_key"`
		ClientCAs  []string `json:"client_cas"`
	} `json:"tls"`

	PKCS11 struct {
		EntityModule struct {
			Path string `json:"path"`
			Pin  string `json:"pin"`
		} `json:"entity_module"`
	} `json:"pkcs11"`

	Endpoint struct {
		MutualAuthentication bool   `json:"mutual_authentication"`
		NoEncryption         bool   `json:"no_encryption"`
		ASLKeyExchangeMethod int    `json:"asl_key_exchange_method"`
		KeylogFile           string `json:"keylog_file"`
	} `json:"endpoint"`

	ASLConfig struct {
		LoggingEnabled          bool `json:"logging_enabled"`
		LogLevel                int  `json:"log_level"`
		SecureElementLogSupport bool `json:"secure_element_log_support"`
	} `json:"asl_config"`

	Database struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		User     string `json:"user"`
		Password string `json:"password"`
		DBName   string `json:"dbname"`
		SSLMode  string `json:"sslmode"`
	} `json:"database"`
}

// Load reads configuration from a JSON file and environment variables
func Load(filepath string, cfg *Config) (*Config, error) {
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

	return cfg, nil
}

// String returns a string representation of the configuration
func (c *Config) String() string {
	b, _ := json.MarshalIndent(c, "", "  ")
	return string(b)
}
