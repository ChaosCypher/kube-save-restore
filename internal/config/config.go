package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

// Config holds the configuration for the application.
type Config struct {
	KubeConfig     string
	Context        string
	BackupDir      string
	RestoreDir     string
	Mode           string
	DryRun         bool
	LogLevel       string
	LogFile        string
	MaxConcurrency int
}

// ParseFlags parses command-line flags and environment variables into a Config struct.
func ParseFlags() *Config {
	config := &Config{}

	flag.StringVar(&config.KubeConfig, "kubeconfig", "", "Path to the kubeconfig file")
	flag.StringVar(&config.Context, "context", "", "Kubernetes context to use")
	flag.StringVar(&config.BackupDir, "backup-dir", "", "Directory to store backups")
	flag.StringVar(&config.RestoreDir, "restore-dir", "", "Directory to restore from")
	flag.StringVar(&config.Mode, "mode", "backup", "Mode of operation: backup or restore")
	flag.BoolVar(&config.DryRun, "dry-run", false, "Perform a dry run without making any changes")
	flag.StringVar(&config.LogLevel, "log-level", "info", "Log level: debug, info, warn, error")
	flag.StringVar(&config.LogFile, "log-file", "", "Path to the log file")
	flag.IntVar(&config.MaxConcurrency, "max-concurrency", 10, "Maximum number of concurrent operations")

	flag.Parse()

	return config
}

// validateConfig validates the configuration values.
func validateConfig(config *Config) error {
	validModes := map[string]bool{"backup": true, "restore": true}
	if !validModes[config.Mode] {
		return fmt.Errorf("invalid mode: %s. Use 'backup' or 'restore'", config.Mode)
	}
	if config.Mode == "restore" && config.RestoreDir == "" {
		return fmt.Errorf("--restore-dir flag is required for restore mode")
	}
	return nil
}

// getEnv retrieves the value of the environment variable named by the key.
// It returns the value, or the specified default value if the variable is not present.
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvAsBool retrieves the value of the environment variable named by the key and parses it as a boolean.
// It returns the boolean value, or the specified default value if the variable is not present or cannot be parsed.
func getEnvAsBool(name string, defaultVal bool) bool {
	valStr := getEnv(name, "")
	if val, err := strconv.ParseBool(valStr); err == nil {
		return val
	}
	return defaultVal
}
