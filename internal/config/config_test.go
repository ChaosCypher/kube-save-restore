package config

import (
	"flag"
	"os"
	"testing"
)

func TestParseFlags(t *testing.T) {
	// Save original command-line arguments and defer restoration
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	// Test cases
	tests := []struct {
		name       string
		args       []string
		envVars    map[string]string
		expectFunc func(*Config) bool
	}{
		{
			name: "Default values",
			args: []string{},
			envVars: map[string]string{
				"KUBECONFIG":   "",
				"KUBE_CONTEXT": "",
				"BACKUP_DIR":   "",
				"RESTORE_DIR":  "",
				"MODE":         "backup",
				"DRY_RUN":      "false",
				"LOG_LEVEL":    "info",
				"LOG_FILE":     "",
			},
			expectFunc: func(config *Config) bool {
				return config.Mode == "backup" && !config.DryRun && config.LogLevel == "info"
			},
		},
		{
			name: "Custom values",
			args: []string{
				"--kubeconfig=/path/to/kubeconfig",
				"--context=my-context",
				"--backup-dir=/path/to/backup",
				"--restore-dir=/path/to/restore",
				"--mode=restore",
				"--dry-run=true",
				"--log-level=debug",
				"--log-file=/path/to/logfile",
			},
			envVars: map[string]string{},
			expectFunc: func(config *Config) bool {
				return config.KubeConfig == "/path/to/kubeconfig" &&
					config.Context == "my-context" &&
					config.BackupDir == "/path/to/backup" &&
					config.RestoreDir == "/path/to/restore" &&
					config.Mode == "restore" &&
					config.DryRun &&
					config.LogLevel == "debug" &&
					config.LogFile == "/path/to/logfile"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// Set command-line arguments
			os.Args = append([]string{"cmd"}, tt.args...)

			// Parse flags
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
			config := ParseFlags()

			// Validate expectations
			if !tt.expectFunc(config) {
				t.Errorf("Test %s failed", tt.name)
			}

			// Unset environment variables
			for key := range tt.envVars {
				os.Unsetenv(key)
			}
		})
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		expectErr bool
	}{
		{
			name: "Valid backup mode",
			config: &Config{
				Mode: "backup",
			},
			expectErr: false,
		},
		{
			name: "Valid restore mode with restore-dir",
			config: &Config{
				Mode:       "restore",
				RestoreDir: "/path/to/restore",
			},
			expectErr: false,
		},
		{
			name: "Invalid mode",
			config: &Config{
				Mode: "invalid",
			},
			expectErr: true,
		},
		{
			name: "Restore mode without restore-dir",
			config: &Config{
				Mode: "restore",
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)
			if (err != nil) != tt.expectErr {
				t.Errorf("Test %s failed: expected error %v, got %v", tt.name, tt.expectErr, err)
			}
		})
	}
}

// TestGetEnv tests the getEnv function.
func TestGetEnv(t *testing.T) {
	// Backup original environment variable and ensure it's restored after the test.
	originalValue, exists := os.LookupEnv("TEST_ENV")
	defer func() {
		if exists {
			os.Setenv("TEST_ENV", originalValue)
		} else {
			os.Unsetenv("TEST_ENV")
		}
	}()

	tests := []struct {
		name         string
		envKey       string
		envValue     string
		envSet       bool
		defaultValue string
		expected     string
	}{
		{
			name:         "Environment variable is set",
			envKey:       "TEST_ENV",
			envValue:     "value1",
			envSet:       true,
			defaultValue: "default1",
			expected:     "value1",
		},
		{
			name:         "Environment variable is not set",
			envKey:       "UNSET_ENV",
			envValue:     "",
			envSet:       false,
			defaultValue: "default2",
			expected:     "default2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envSet {
				os.Setenv(tt.envKey, tt.envValue)
			} else {
				os.Unsetenv(tt.envKey)
			}

			result := getEnv(tt.envKey, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnv(%s, %s) = %s; want %s", tt.envKey, tt.defaultValue, result, tt.expected)
			}
		})
	}
}

// TestGetEnvAsBool tests the getEnvAsBool function.
func TestGetEnvAsBool(t *testing.T) {
	// Backup original environment variable and ensure it's restored after the test.
	originalValue, exists := os.LookupEnv("TEST_BOOL_ENV")
	defer func() {
		if exists {
			os.Setenv("TEST_BOOL_ENV", originalValue)
		} else {
			os.Unsetenv("TEST_BOOL_ENV")
		}
	}()

	tests := []struct {
		name         string
		envKey       string
		envValue     string
		envSet       bool
		defaultValue bool
		expected     bool
	}{
		{
			name:         "Environment variable is true",
			envKey:       "TEST_BOOL_ENV",
			envValue:     "true",
			envSet:       true,
			defaultValue: false,
			expected:     true,
		},
		{
			name:         "Environment variable is false",
			envKey:       "TEST_BOOL_ENV",
			envValue:     "false",
			envSet:       true,
			defaultValue: true,
			expected:     false,
		},
		{
			name:         "Environment variable is not set, use default true",
			envKey:       "UNSET_BOOL_ENV",
			envValue:     "",
			envSet:       false,
			defaultValue: true,
			expected:     true,
		},
		{
			name:         "Environment variable is not set, use default false",
			envKey:       "UNSET_BOOL_ENV",
			envValue:     "",
			envSet:       false,
			defaultValue: false,
			expected:     false,
		},
		{
			name:         "Environment variable has invalid value, use default",
			envKey:       "TEST_BOOL_ENV",
			envValue:     "invalid",
			envSet:       true,
			defaultValue: true,
			expected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envSet {
				os.Setenv(tt.envKey, tt.envValue)
			} else {
				os.Unsetenv(tt.envKey)
			}

			result := getEnvAsBool(tt.envKey, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnvAsBool(%s, %v) = %v; want %v", tt.envKey, tt.defaultValue, result, tt.expected)
			}
		})
	}
}
