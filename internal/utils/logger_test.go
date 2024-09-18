package utils

import (
	"bytes"
	"fmt"
	"k8s-backup-restore/internal/config"
	"os"
	"strings"
	"testing"
)

// TestNewLogger tests the creation of a new Logger instance
func TestNewLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf, INFO)

	if logger.level != INFO {
		t.Errorf("Expected log level %v, got %v", INFO, logger.level)
	}
	if logger.output != &buf {
		t.Errorf("Expected output %v, got %v", &buf, logger.output)
	}
}

// TestSetupLogger tests the setup of the logger based on configuration
func TestSetupLogger(t *testing.T) {
	cfg := &config.Config{LogFile: "", LogLevel: "debug"}
	logger := SetupLogger(cfg)

	if logger.level != DEBUG {
		t.Errorf("Expected log level %v, got %v", DEBUG, logger.level)
	}
	if logger.output != os.Stdout {
		t.Errorf("Expected output %v, got %v", os.Stdout, logger.output)
	}
}

// TestSetupLoggerWithFile tests the setup of the logger with a log file
func TestSetupLoggerWithFile(t *testing.T) {
	logFile := "test.log"
	defer os.Remove(logFile) // Clean up

	cfg := &config.Config{LogFile: logFile, LogLevel: "info"}
	logger := SetupLogger(cfg)

	if logger.level != INFO {
		t.Errorf("Expected log level %v, got %v", INFO, logger.level)
	}
	if logger.output == os.Stdout {
		t.Errorf("Expected output to be a file, got %v", logger.output)
	}
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Errorf("Expected log file %s to be created", logFile)
	}
}

// TestLoggerEdgeCases tests edge cases for the logger
func TestLoggerEdgeCases(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf, DEBUG)

	// Test with empty log message
	logger.Debug("")
	if !bytes.Contains(buf.Bytes(), []byte("DEBUG: ")) {
		t.Errorf("Expected empty debug log message not found")
	}

	// Test with very long log message
	longMessage := string(make([]byte, 1024*1024)) // 1MB log message
	logger.Debug(longMessage)
	if !bytes.Contains(buf.Bytes(), []byte(longMessage)) {
		t.Errorf("Expected long debug log message not found")
	}
}

// TestLogger_Close tests the Close method of the Logger.
func TestLogger_Close(t *testing.T) {
	t.Run("Close with logFile", func(t *testing.T) {
		// Create a temporary file to simulate logFile
		tmpFile, err := os.CreateTemp("", "logger_test.log")
		if err != nil {
			t.Fatalf("Failed to create temporary log file: %v", err)
		}
		defer os.Remove(tmpFile.Name())

		// Initialize Logger with the temporary file
		logger := &Logger{
			logFile: tmpFile,
		}

		// Call Close method
		logger.Close()

		// Attempt to write to the closed file to ensure it's closed
		_, err = logger.logFile.WriteString("test")
		if err == nil {
			t.Error("Expected error when writing to closed file, but got none")
		}
	})

	t.Run("Close without logFile", func(t *testing.T) {
		// Initialize Logger without a logFile
		logger := &Logger{
			logFile: nil,
		}

		// Ensure that calling Close does not panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Close() panicked when logFile is nil: %v", r)
			}
		}()

		// Call Close method
		logger.Close()
	})
}

// TestParseLogLevel tests the parseLogLevel function to ensure it correctly parses string log levels.
func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected LogLevel
	}{
		{"debug", DEBUG},
		{"info", INFO},
		{"warn", WARN},
		{"error", ERROR},
		{"UNKNOWN", INFO}, // Default case
		{"", INFO},        // Empty string
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("Input_%s", tt.input), func(t *testing.T) {
			result := parseLogLevel(tt.input)
			if result != tt.expected {
				t.Errorf("parseLogLevel(%q) = %v; want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestLogger_Debug tests the Debug method of the Logger.
func TestLogger_Debug(t *testing.T) {
	tests := []struct {
		name          string
		logLevel      LogLevel
		shouldLog     bool
		testMessage   string
		expectedPrefix string
	}{
		{
			name:          "Level_DEBUG_ShouldLog",
			logLevel:      DEBUG,
			shouldLog:     true,
			testMessage:   "This is a debug message",
			expectedPrefix: "DEBUG: ",
		},
		{
			name:          "Level_INFO_ShouldNotLog",
			logLevel:      INFO,
			shouldLog:     false,
			testMessage:   "This debug message should not appear",
			expectedPrefix: "DEBUG: ",
		},
		{
			name:          "Level_WARN_ShouldNotLog",
			logLevel:      WARN,
			shouldLog:     false,
			testMessage:   "Another debug message that should not be logged",
			expectedPrefix: "DEBUG: ",
		},
		{
			name:          "Level_ERROR_ShouldNotLog",
			logLevel:      ERROR,
			shouldLog:     false,
			testMessage:   "Error level should suppress debug messages",
			expectedPrefix: "DEBUG: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewLogger(&buf, tt.logLevel)
			logger.Debug(tt.testMessage)

			logOutput := buf.String()
			if tt.shouldLog {
				if !strings.Contains(logOutput, tt.expectedPrefix) || !strings.Contains(logOutput, tt.testMessage) {
					t.Errorf("Expected log output to contain '%s' and message '%s'. Got: %s", tt.expectedPrefix, tt.testMessage, logOutput)
				}
			} else {
				if strings.Contains(logOutput, tt.expectedPrefix) {
					t.Errorf("Did not expect DEBUG log output when level is %v. Got: %s", tt.logLevel, logOutput)
				}
			}
		})
	}
}

func TestLogger_Info(t *testing.T) {
	tests := []struct {
		name           string
		logLevel       LogLevel
		shouldLog      bool
		testMessage    string
		expectedPrefix string
	}{
		{
			name:           "Level_DEBUG_ShouldLog",
			logLevel:       DEBUG,
			shouldLog:      true,
			testMessage:    "This is an info message",
			expectedPrefix: "INFO: ",
		},
		{
			name:           "Level_INFO_ShouldLog",
			logLevel:       INFO,
			shouldLog:      true,
			testMessage:    "This is another info message",
			expectedPrefix: "INFO: ",
		},
		{
			name:           "Level_WARN_ShouldNotLog",
			logLevel:       WARN,
			shouldLog:      false,
			testMessage:    "This info message should not appear",
			expectedPrefix: "INFO: ",
		},
		{
			name:           "Level_ERROR_ShouldNotLog",
			logLevel:       ERROR,
			shouldLog:      false,
			testMessage:    "Error level should suppress info messages",
			expectedPrefix: "INFO: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewLogger(&buf, tt.logLevel)
			logger.Info(tt.testMessage)

			logOutput := buf.String()
			if tt.shouldLog {
				if !strings.Contains(logOutput, tt.expectedPrefix) || !strings.Contains(logOutput, tt.testMessage) {
					t.Errorf("Expected log output to contain '%s' and message '%s'. Got: %s", tt.expectedPrefix, tt.testMessage, logOutput)
				}
			} else {
				if strings.Contains(logOutput, tt.expectedPrefix) {
					t.Errorf("Did not expect INFO log output when level is %v. Got: %s", tt.logLevel, logOutput)
				}
			}
		})
	}
}

// TestLogger_Warn tests the Warn method of the Logger.
func TestLogger_Warn(t *testing.T) {
	tests := []struct {
		name           string
		logLevel       LogLevel
		shouldLog      bool
		testMessage    string
		expectedPrefix string
	}{
		{
			name:           "Level_DEBUG_ShouldLog",
			logLevel:       DEBUG,
			shouldLog:      true,
			testMessage:    "This is a warning message",
			expectedPrefix: "WARN: ",
		},
		{
			name:           "Level_INFO_ShouldLog",
			logLevel:       INFO,
			shouldLog:      true,
			testMessage:    "This is another warning message",
			expectedPrefix: "WARN: ",
		},
		{
			name:           "Level_WARN_ShouldLog",
			logLevel:       WARN,
			shouldLog:      true,
			testMessage:    "Warning level should log warnings",
			expectedPrefix: "WARN: ",
		},
		{
			name:           "Level_ERROR_ShouldNotLog",
			logLevel:       ERROR,
			shouldLog:      false,
			testMessage:    "Error level should suppress warn messages",
			expectedPrefix: "WARN: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewLogger(&buf, tt.logLevel)
			logger.Warn(tt.testMessage)

			logOutput := buf.String()
			if tt.shouldLog {
				if !strings.Contains(logOutput, tt.expectedPrefix) || !strings.Contains(logOutput, tt.testMessage) {
					t.Errorf("Expected log output to contain '%s' and message '%s'. Got: %s", tt.expectedPrefix, tt.testMessage, logOutput)
				}
			} else {
				if strings.Contains(logOutput, tt.expectedPrefix) {
					t.Errorf("Did not expect WARN log output when level is %v. Got: %s", tt.logLevel, logOutput)
				}
			}
		})
	}
}

// TestLogger_Error tests the Error method of the Logger.
func TestLogger_Error(t *testing.T) {
	tests := []struct {
		name           string
		logLevel       LogLevel
		shouldLog      bool
		testMessage    string
		expectedPrefix string
	}{
		{
			name:           "Level_DEBUG_ShouldLog",
			logLevel:       DEBUG,
			shouldLog:      true,
			testMessage:    "This is an error message",
			expectedPrefix: "ERROR: ",
		},
		{
			name:           "Level_INFO_ShouldLog",
			logLevel:       INFO,
			shouldLog:      true,
			testMessage:    "This is another error message",
			expectedPrefix: "ERROR: ",
		},
		{
			name:           "Level_WARN_ShouldLog",
			logLevel:       WARN,
			shouldLog:      true,
			testMessage:    "Warn level should allow error messages",
			expectedPrefix: "ERROR: ",
		},
		{
			name:           "Level_ERROR_ShouldLog",
			logLevel:       ERROR,
			shouldLog:      true,
			testMessage:    "Error level should log error messages",
			expectedPrefix: "ERROR: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewLogger(&buf, tt.logLevel)
			logger.Error(tt.testMessage)

			logOutput := buf.String()
			if tt.shouldLog {
				if !strings.Contains(logOutput, tt.expectedPrefix) || !strings.Contains(logOutput, tt.testMessage) {
					t.Errorf("Expected log output to contain '%s' and message '%s'. Got: %s", tt.expectedPrefix, tt.testMessage, logOutput)
				}
			} else {
				if strings.Contains(logOutput, tt.expectedPrefix) {
					t.Errorf("Did not expect ERROR log output when level is %v. Got: %s", tt.logLevel, logOutput)
				}
			}
		})
	}
}

// TestLogger_Debugf tests the Debugf method of the Logger.
func TestLogger_Debugf(t *testing.T) {
	tests := []struct {
		name           string
		logLevel       LogLevel
		shouldLog      bool
		format         string
		args           []interface{}
		expectedPrefix string
	}{
		{
			name:           "Level_DEBUG_ShouldLog",
			logLevel:       DEBUG,
			shouldLog:      true,
			format:         "Debugf message: %s",
			args:           []interface{}{"test"},
			expectedPrefix: "DEBUG: ",
		},
		{
			name:           "Level_INFO_ShouldNotLog",
			logLevel:       INFO,
			shouldLog:      false,
			format:         "This debugf message should not appear: %d",
			args:           []interface{}{123},
			expectedPrefix: "DEBUG: ",
		},
		{
			name:           "Level_WARN_ShouldNotLog",
			logLevel:       WARN,
			shouldLog:      false,
			format:         "Another debugf message: %f",
			args:           []interface{}{3.14},
			expectedPrefix: "DEBUG: ",
		},
		{
			name:           "Level_ERROR_ShouldNotLog",
			logLevel:       ERROR,
			shouldLog:      false,
			format:         "Error level should suppress debugf: %v",
			args:           []interface{}{nil},
			expectedPrefix: "DEBUG: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewLogger(&buf, tt.logLevel)
			logger.Debugf(tt.format, tt.args...)

			logOutput := buf.String()
			if tt.shouldLog {
				expectedMessage := fmt.Sprintf(tt.format, tt.args...)
				if !strings.Contains(logOutput, tt.expectedPrefix) || !strings.Contains(logOutput, expectedMessage) {
					t.Errorf("Expected log output to contain '%s' and message '%s'. Got: %s", tt.expectedPrefix, expectedMessage, logOutput)
				}
			} else {
				if strings.Contains(logOutput, tt.expectedPrefix) {
					t.Errorf("Did not expect DEBUGf log output when level is %v. Got: %s", tt.logLevel, logOutput)
				}
			}
		})
	}
}

// TestLogger_Infof tests the Infof method of the Logger.
func TestLogger_Infof(t *testing.T) {
	tests := []struct {
		name           string
		logLevel       LogLevel
		shouldLog      bool
		format         string
		args           []interface{}
		expectedPrefix string
	}{
		{
			name:           "Level_DEBUG_ShouldLog",
			logLevel:       DEBUG,
			shouldLog:      true,
			format:         "Infof message: %s",
			args:           []interface{}{"test"},
			expectedPrefix: "INFO: ",
		},
		{
			name:           "Level_INFO_ShouldLog",
			logLevel:       INFO,
			shouldLog:      true,
			format:         "Another Infof message: %d",
			args:           []interface{}{456},
			expectedPrefix: "INFO: ",
		},
		{
			name:           "Level_WARN_ShouldNotLog",
			logLevel:       WARN,
			shouldLog:      false,
			format:         "This Infof message should not appear: %f",
			args:           []interface{}{6.28},
			expectedPrefix: "INFO: ",
		},
		{
			name:           "Level_ERROR_ShouldNotLog",
			logLevel:       ERROR,
			shouldLog:      false,
			format:         "Error level should suppress Infof: %v",
			args:           []interface{}{nil},
			expectedPrefix: "INFO: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewLogger(&buf, tt.logLevel)
			logger.Infof(tt.format, tt.args...)

			logOutput := buf.String()
			if tt.shouldLog {
				expectedMessage := fmt.Sprintf(tt.format, tt.args...)
				if !strings.Contains(logOutput, tt.expectedPrefix) || !strings.Contains(logOutput, expectedMessage) {
					t.Errorf("Expected log output to contain '%s' and message '%s'. Got: %s", tt.expectedPrefix, expectedMessage, logOutput)
				}
			} else {
				if strings.Contains(logOutput, tt.expectedPrefix) {
					t.Errorf("Did not expect INFOf log output when level is %v. Got: %s", tt.logLevel, logOutput)
				}
			}
		})
	}
}

// TestLogger_Warnf tests the Warnf method of the Logger.
func TestLogger_Warnf(t *testing.T) {
	tests := []struct {
		name           string
		logLevel       LogLevel
		shouldLog      bool
		format         string
		args           []interface{}
		expectedPrefix string
	}{
		{
			name:           "Level_DEBUG_ShouldLog",
			logLevel:       DEBUG,
			shouldLog:      true,
			format:         "Warnf message: %s",
			args:           []interface{}{"test"},
			expectedPrefix: "WARN: ",
		},
		{
			name:           "Level_INFO_ShouldLog",
			logLevel:       INFO,
			shouldLog:      true,
			format:         "Another Warnf message: %d",
			args:           []interface{}{789},
			expectedPrefix: "WARN: ",
		},
		{
			name:           "Level_WARN_ShouldLog",
			logLevel:       WARN,
			shouldLog:      true,
			format:         "Warn level should log Warnf messages: %f",
			args:           []interface{}{9.42},
			expectedPrefix: "WARN: ",
		},
		{
			name:           "Level_ERROR_ShouldNotLog",
			logLevel:       ERROR,
			shouldLog:      false,
			format:         "Error level should suppress Warnf: %v",
			args:           []interface{}{nil},
			expectedPrefix: "WARN: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewLogger(&buf, tt.logLevel)
			logger.Warnf(tt.format, tt.args...)

			logOutput := buf.String()
			if tt.shouldLog {
				expectedMessage := fmt.Sprintf(tt.format, tt.args...)
				if !strings.Contains(logOutput, tt.expectedPrefix) || !strings.Contains(logOutput, expectedMessage) {
					t.Errorf("Expected log output to contain '%s' and message '%s'. Got: %s", tt.expectedPrefix, expectedMessage, logOutput)
				}
			} else {
				if strings.Contains(logOutput, tt.expectedPrefix) {
					t.Errorf("Did not expect WARNf log output when level is %v. Got: %s", tt.logLevel, logOutput)
				}
			}
		})
	}
}

// TestLogger_Errorf tests the Errorf method of the Logger.
func TestLogger_Errorf(t *testing.T) {
	tests := []struct {
		name           string
		logLevel       LogLevel
		shouldLog      bool
		format         string
		args           []interface{}
		expectedPrefix string
	}{
		{
			name:           "Level_DEBUG_ShouldLog",
			logLevel:       DEBUG,
			shouldLog:      true,
			format:         "Errorf message: %s",
			args:           []interface{}{"test"},
			expectedPrefix: "ERROR: ",
		},
		{
			name:           "Level_INFO_ShouldLog",
			logLevel:       INFO,
			shouldLog:      true,
			format:         "Another Errorf message: %d",
			args:           []interface{}{101112},
			expectedPrefix: "ERROR: ",
		},
		{
			name:           "Level_WARN_ShouldLog",
			logLevel:       WARN,
			shouldLog:      true,
			format:         "Warn level should log Errorf messages: %f",
			args:           []interface{}{12.56},
			expectedPrefix: "ERROR: ",
		},
		{
			name:           "Level_ERROR_ShouldLog",
			logLevel:       ERROR,
			shouldLog:      true,
			format:         "Error level should log Errorf messages: %v",
			args:           []interface{}{nil},
			expectedPrefix: "ERROR: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewLogger(&buf, tt.logLevel)
			logger.Errorf(tt.format, tt.args...)

			logOutput := buf.String()
			if tt.shouldLog {
				expectedMessage := fmt.Sprintf(tt.format, tt.args...)
				if !strings.Contains(logOutput, tt.expectedPrefix) || !strings.Contains(logOutput, expectedMessage) {
					t.Errorf("Expected log output to contain '%s' and message '%s'. Got: %s", tt.expectedPrefix, expectedMessage, logOutput)
				}
			} else {
				if strings.Contains(logOutput, tt.expectedPrefix) {
					t.Errorf("Did not expect ERRORf log output when level is %v. Got: %s", tt.logLevel, logOutput)
				}
			}
		})
	}
}
