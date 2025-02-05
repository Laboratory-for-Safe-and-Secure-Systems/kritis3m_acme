package config

import (
	"flag"
	"fmt"
	"os"
)

// Flags holds all the command-line flags for the application
type Flags struct {
	ConfigPath string
	Debug      bool
	Version    bool
}

var (
	// GlobalFlags stores the parsed command-line flags
	GlobalFlags = &Flags{}

	// Version information (to be set at build time)
	Version   = "development"
	GitCommit = "unknown"
	BuildTime = "unknown"
)

// ParseFlags parses command-line flags and returns any error encountered
func ParseFlags() error {
	// Define flags
	flag.StringVar(&GlobalFlags.ConfigPath, "config", "", "path to configuration file")
	flag.BoolVar(&GlobalFlags.Debug, "debug", false, "enable debug mode")
	flag.BoolVar(&GlobalFlags.Version, "version", false, "print version information")

	// Custom usage message
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Environment variables:\n")
		fmt.Fprintf(os.Stderr, "  SERVER_PORT       Override server port\n")
		fmt.Fprintf(os.Stderr, "  SERVER_HOST       Override server host\n")
	}

	// Parse flags
	flag.Parse()

	// Handle version flag
	if GlobalFlags.Version {
		fmt.Printf("Version:    %s\n", Version)
		fmt.Printf("Git commit: %s\n", GitCommit)
		fmt.Printf("Built:      %s\n", BuildTime)
		os.Exit(0)
	}

	return nil
}

// LoadConfig loads the configuration using the provided flags
func LoadConfig() (*Config, error) {
	// Parse command-line flags
	if err := ParseFlags(); err != nil {
		return nil, fmt.Errorf("error parsing flags: %w", err)
	}

	// Load configuration
	cfg, err := Load(GlobalFlags.ConfigPath, &Config{})
	if err != nil {
		return nil, fmt.Errorf("error loading config: %w", err)
	}

	// Apply debug mode if set
	if GlobalFlags.Debug {
		fmt.Println("Debug mode enabled")
		// Add any debug-specific configuration here
	}

	return cfg, nil
}

// GetNonFlagArgs returns all non-flag command-line arguments
func GetNonFlagArgs() []string {
	return flag.Args()
}
