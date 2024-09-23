package logger

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/chaoscypher/k8s-backup-restore/internal/config"
)

// String returns the string representation of the LogLevel.
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Helper function to create a new Logger with a buffer and specified log level.
func createTestLogger(t *testing.T, level LogLevel) (*Logger, *bytes.Buffer) {
	t.Helper()
	var buf bytes.Buffer
	logger := NewLogger(&buf, level)
	return logger, &buf
}

// Helper function to setup Logger based on config.
func setupLoggerFromConfig(t *testing.T, cfg *config.Config) *Logger {
	t.Helper()
	logger, ok := SetupLogger(cfg).(*Logger)
	if !ok {
		t.Fatalf("expected *Logger, got %T", logger)
	}
	return logger
}

// Helper function to perform logging and verify the output.
func logAndVerify(t *testing.T, logger *Logger, buf *bytes.Buffer, logFunc func(*Logger, string, ...interface{}), shouldLog bool, expectedPrefix string, message string, args ...interface{}) {
	t.Helper()
	logFunc(logger, message, args...)
	logOutput := buf.String()

	if shouldLog {
		expectedMessage := fmt.Sprintf(message, args...)
		if !strings.Contains(logOutput, expectedPrefix) || !strings.Contains(logOutput, expectedMessage) {
			t.Errorf("Expected log output to contain '%s' and message '%s'. Got: %s", expectedPrefix, expectedMessage, logOutput)
		}
	} else if strings.Contains(logOutput, expectedPrefix) {
		t.Errorf("Did not expect log output with prefix '%s'. Got: %s", expectedPrefix, logOutput)
	}
}

// TestNewLogger tests the creation of a new Logger instance.
func TestNewLogger(t *testing.T) {
	logger, buf := createTestLogger(t, INFO)

	if logger.Level != INFO {
		t.Errorf("Expected log level %v, got %v", INFO, logger.Level)
	}
	if logger.Output != buf {
		t.Errorf("Expected output %v, got %v", buf, logger.Output)
	}
}

// TestSetupLogger tests the setup of the logger based on configuration.
func TestSetupLogger(t *testing.T) {
	cfg := &config.Config{LogFile: "", LogLevel: "debug"}
	logger := setupLoggerFromConfig(t, cfg)

	if logger.Level != DEBUG {
		t.Errorf("Expected log level %v, got %v", DEBUG, logger.Level)
	}
	if logger.Output != os.Stdout {
		t.Errorf("Expected output %v, got %v", os.Stdout, logger.Output)
	}
}

// TestSetupLoggerWithFile tests the setup of the logger with a log file.
func TestSetupLoggerWithFile(t *testing.T) {
	logFile := "test.log"
	defer os.Remove(logFile) // Clean up

	cfg := &config.Config{LogFile: logFile, LogLevel: "info"}
	logger := setupLoggerFromConfig(t, cfg)

	if logger.Level != INFO {
		t.Errorf("Expected log level %v, got %v", INFO, logger.Level)
	}
	if logger.Output == os.Stdout {
		t.Errorf("Expected output to be a file, got %v", logger.Output)
	}
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Errorf("Expected log file %s to be created", logFile)
	}
}

// TestLoggerEdgeCases tests edge cases for the logger.
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
			LogFile: tmpFile,
		}

		// Call Close method
		logger.Close()

		// Attempt to write to the closed file to ensure it's closed
		_, err = logger.LogFile.WriteString("test")
		if err == nil {
			t.Error("Expected error when writing to closed file, but got none")
		}
	})

	t.Run("Close without logFile", func(t *testing.T) {
		// Initialize Logger without a logFile
		logger := &Logger{
			LogFile: nil,
		}

		// Ensure that calling Close does not panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Close() panicked when LogFile is nil: %v", r)
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

// TestLoggerMethods tests the various logging methods of the Logger.
func TestLoggerMethods(t *testing.T) {
	tests := []struct {
		method         string
		logLevel       LogLevel
		shouldLog      bool
		logFunc        func(*Logger, string)
		expectedPrefix string
		message        string
	}{
		// Debug
		{"Debug", DEBUG, true, func(l *Logger, msg string) { l.Debug(msg) }, "DEBUG: ", "This is a debug message"},
		{"Debug", INFO, false, func(l *Logger, msg string) { l.Debug(msg) }, "DEBUG: ", "This debug message should not appear"},
		{"Debug", WARN, false, func(l *Logger, msg string) { l.Debug(msg) }, "DEBUG: ", "Another debug message that should not be logged"},
		{"Debug", ERROR, false, func(l *Logger, msg string) { l.Debug(msg) }, "DEBUG: ", "Error level should suppress debug messages"},

		// Info
		{"Info", DEBUG, true, func(l *Logger, msg string) { l.Info(msg) }, "INFO: ", "This is an info message"},
		{"Info", INFO, true, func(l *Logger, msg string) { l.Info(msg) }, "INFO: ", "This is another info message"},
		{"Info", WARN, false, func(l *Logger, msg string) { l.Info(msg) }, "INFO: ", "This info message should not appear"},
		{"Info", ERROR, false, func(l *Logger, msg string) { l.Info(msg) }, "INFO: ", "Error level should suppress info messages"},

		// Warn
		{"Warn", DEBUG, true, func(l *Logger, msg string) { l.Warn(msg) }, "WARN: ", "This is a warning message"},
		{"Warn", INFO, true, func(l *Logger, msg string) { l.Warn(msg) }, "WARN: ", "This is another warning message"},
		{"Warn", WARN, true, func(l *Logger, msg string) { l.Warn(msg) }, "WARN: ", "Warning level should log warnings"},
		{"Warn", ERROR, false, func(l *Logger, msg string) { l.Warn(msg) }, "WARN: ", "Error level should suppress warn messages"},

		// Error
		{"Error", DEBUG, true, func(l *Logger, msg string) { l.Error(msg) }, "ERROR: ", "This is an error message"},
		{"Error", INFO, true, func(l *Logger, msg string) { l.Error(msg) }, "ERROR: ", "This is another error message"},
		{"Error", WARN, true, func(l *Logger, msg string) { l.Error(msg) }, "ERROR: ", "Warn level should allow error messages"},
		{"Error", ERROR, true, func(l *Logger, msg string) { l.Error(msg) }, "ERROR: ", "Error level should log error messages"},
	}

	for _, tt := range tests {
		testName := fmt.Sprintf("%s_%s", tt.method, tt.logLevel.String())
		t.Run(testName, func(t *testing.T) {
			logger, buf := createTestLogger(t, tt.logLevel)
			tt.logFunc(logger, tt.message)

			logOutput := buf.String()
			if tt.shouldLog {
				if !strings.Contains(logOutput, tt.expectedPrefix) || !strings.Contains(logOutput, tt.message) {
					t.Errorf("Expected log output to contain '%s' and message '%s'. Got: %s", tt.expectedPrefix, tt.message, logOutput)
				}
			} else {
				if strings.Contains(logOutput, tt.expectedPrefix) {
					t.Errorf("Did not expect log output with prefix '%s'. Got: %s", tt.expectedPrefix, logOutput)
				}
			}
		})
	}
}

// TestLoggerFormattedMethods tests the formatted logging methods of the Logger.
func TestLoggerFormattedMethods(t *testing.T) {
	tests := []struct {
		method         string
		logLevel       LogLevel
		shouldLog      bool
		logFunc        func(*Logger, string, ...interface{})
		expectedPrefix string
		format         string
		args           []interface{}
	}{
		// Debugf
		{"Debugf", DEBUG, true, func(l *Logger, fmtStr string, args ...interface{}) { l.Debugf(fmtStr, args...) }, "DEBUG: ", "Debugf message: %s", []interface{}{"test"}},
		{"Debugf", INFO, false, func(l *Logger, fmtStr string, args ...interface{}) { l.Debugf(fmtStr, args...) }, "DEBUG: ", "This debugf message should not appear: %d", []interface{}{123}},
		{"Debugf", WARN, false, func(l *Logger, fmtStr string, args ...interface{}) { l.Debugf(fmtStr, args...) }, "DEBUG: ", "Another debugf message: %f", []interface{}{3.14}},
		{"Debugf", ERROR, false, func(l *Logger, fmtStr string, args ...interface{}) { l.Debugf(fmtStr, args...) }, "DEBUG: ", "Error level should suppress debugf: %v", []interface{}{nil}},

		// Infof
		{"Infof", DEBUG, true, func(l *Logger, fmtStr string, args ...interface{}) { l.Infof(fmtStr, args...) }, "INFO: ", "Infof message: %s", []interface{}{"test"}},
		{"Infof", INFO, true, func(l *Logger, fmtStr string, args ...interface{}) { l.Infof(fmtStr, args...) }, "INFO: ", "Another Infof message: %d", []interface{}{456}},
		{"Infof", WARN, false, func(l *Logger, fmtStr string, args ...interface{}) { l.Infof(fmtStr, args...) }, "INFO: ", "This Infof message should not appear: %f", []interface{}{6.28}},
		{"Infof", ERROR, false, func(l *Logger, fmtStr string, args ...interface{}) { l.Infof(fmtStr, args...) }, "INFO: ", "Error level should suppress Infof: %v", []interface{}{nil}},

		// Warnf
		{"Warnf", DEBUG, true, func(l *Logger, fmtStr string, args ...interface{}) { l.Warnf(fmtStr, args...) }, "WARN: ", "Warnf message: %s", []interface{}{"test"}},
		{"Warnf", INFO, true, func(l *Logger, fmtStr string, args ...interface{}) { l.Warnf(fmtStr, args...) }, "WARN: ", "Another Warnf message: %d", []interface{}{789}},
		{"Warnf", WARN, true, func(l *Logger, fmtStr string, args ...interface{}) { l.Warnf(fmtStr, args...) }, "WARN: ", "Warn level should log Warnf messages: %f", []interface{}{9.42}},
		{"Warnf", ERROR, false, func(l *Logger, fmtStr string, args ...interface{}) { l.Warnf(fmtStr, args...) }, "WARN: ", "Error level should suppress Warnf: %v", []interface{}{nil}},

		// Errorf
		{"Errorf", DEBUG, true, func(l *Logger, fmtStr string, args ...interface{}) { l.Errorf(fmtStr, args...) }, "ERROR: ", "Errorf message: %s", []interface{}{"test"}},
		{"Errorf", INFO, true, func(l *Logger, fmtStr string, args ...interface{}) { l.Errorf(fmtStr, args...) }, "ERROR: ", "Another Errorf message: %d", []interface{}{101112}},
		{"Errorf", WARN, true, func(l *Logger, fmtStr string, args ...interface{}) { l.Errorf(fmtStr, args...) }, "ERROR: ", "Warn level should log Errorf messages: %f", []interface{}{12.56}},
		{"Errorf", ERROR, true, func(l *Logger, fmtStr string, args ...interface{}) { l.Errorf(fmtStr, args...) }, "ERROR: ", "Error level should log Errorf messages: %v", []interface{}{nil}},
	}

	for _, tt := range tests {
		testName := fmt.Sprintf("%sf_%s", tt.method, tt.logLevel)
		t.Run(testName, func(t *testing.T) {
			logger, buf := createTestLogger(t, tt.logLevel)
			logAndVerify(t, logger, buf, tt.logFunc, tt.shouldLog, tt.expectedPrefix, tt.format, tt.args...)

			logOutput := buf.String()
			if tt.shouldLog {
				expectedMessage := fmt.Sprintf(tt.format, tt.args...)
				if !strings.Contains(logOutput, tt.expectedPrefix) || !strings.Contains(logOutput, expectedMessage) {
					t.Errorf("Expected log output to contain '%s' and message '%s'. Got: %s", tt.expectedPrefix, expectedMessage, logOutput)
				}
			} else if strings.Contains(logOutput, tt.expectedPrefix) {
				t.Errorf("Did not expect %sf log output when level is %v. Got: %s", tt.method, tt.logLevel, logOutput)
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
			} else if strings.Contains(logOutput, tt.expectedPrefix) {
				t.Errorf("Did not expect INFOf log output when level is %v. Got: %s", tt.logLevel, logOutput)
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
			} else if strings.Contains(logOutput, tt.expectedPrefix) {
				t.Errorf("Did not expect WARNf log output when level is %v. Got: %s", tt.logLevel, logOutput)
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
			} else if strings.Contains(logOutput, tt.expectedPrefix) {
				t.Errorf("Did not expect ERRORf log output when level is %v. Got: %s", tt.logLevel, logOutput)
			}
		})
	}
}
