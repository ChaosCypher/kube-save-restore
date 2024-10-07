package logger

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/chaoscypher/kube-save-restore/internal/config"
)

// TestLogLevels verifies that each log level function correctly logs messages
func TestLogLevels(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf, DEBUG)

	tests := []struct {
		level   LogLevel
		logFunc func(...interface{})
		message string
	}{
		{DEBUG, logger.Debug, "Debug message"},
		{INFO, logger.Info, "Info message"},
		{WARN, logger.Warn, "Warn message"},
		{ERROR, logger.Error, "Error message"},
	}

	for _, tt := range tests {
		buf.Reset()
		tt.logFunc(tt.message)
		if !bytes.Contains(buf.Bytes(), []byte(tt.message)) {
			t.Errorf("Expected %s to be logged, but it wasn't", tt.message)
		}
	}
}

// TestLogfLevels verifies that each formatted log level function correctly logs messages
func TestLogfLevels(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf, DEBUG)

	tests := []struct {
		level   LogLevel
		logFunc func(string, ...interface{})
		format  string
		args    []interface{}
		message string
	}{
		{DEBUG, logger.Debugf, "Debug %s", []interface{}{"message"}, "Debug message"},
		{INFO, logger.Infof, "Info %s", []interface{}{"message"}, "Info message"},
		{WARN, logger.Warnf, "Warn %s", []interface{}{"message"}, "Warn message"},
		{ERROR, logger.Errorf, "Error %s", []interface{}{"message"}, "Error message"},
	}

	for _, tt := range tests {
		buf.Reset()
		tt.logFunc(tt.format, tt.args...)
		if !bytes.Contains(buf.Bytes(), []byte(tt.message)) {
			t.Errorf("Expected %s to be logged, but it wasn't", tt.message)
		}
	}
}

// TestSetupLogger verifies the logger setup with different configurations
func TestSetupLogger(t *testing.T) {
	// Test case 1: Logger with stdout
	t.Run("Logger with stdout", func(t *testing.T) {
		cfg := &config.Config{
			LogLevel: "debug",
			LogFile:  "",
		}

		logger := SetupLogger(cfg)
		if logger == nil {
			t.Fatal("Expected logger to be initialized, but got nil")
		}

		// Test if the logger logs at the correct level
		var buf bytes.Buffer
		logger = NewLogger(&buf, DEBUG)
		logger.Debug("Test message")
		if !bytes.Contains(buf.Bytes(), []byte("Test message")) {
			t.Errorf("Expected 'Test message' to be logged, but it wasn't")
		}
	})

	// Test case 2: Logger with file output
	t.Run("Logger with file output", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "test_log_*.log")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tempFile.Name())

		cfg := &config.Config{
			LogLevel: "debug",
			LogFile:  tempFile.Name(),
		}

		logger := SetupLogger(cfg)
		if logger == nil {
			t.Fatal("Expected logger to be initialized, but got nil")
		}

		// Test if the logger writes to the file
		testMessage := "File logger test message"
		logger.Debug(testMessage)
		logger.Close()

		content, err := os.ReadFile(tempFile.Name())
		if err != nil {
			t.Fatalf("Failed to read log file: %v", err)
		}

		if !bytes.Contains(content, []byte(testMessage)) {
			t.Errorf("Expected log file to contain %q, but it didn't", testMessage)
		}
	})
}

// TestParseLogLevel verifies that log levels are correctly parsed from strings
func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected LogLevel
	}{
		{"debug", DEBUG},
		{"info", INFO},
		{"warn", WARN},
		{"error", ERROR},
		{"unknown", INFO},
	}

	for _, tt := range tests {
		result := parseLogLevel(tt.input)
		if result != tt.expected {
			t.Errorf("Expected %v for input %s, but got %v", tt.expected, tt.input, result)
		}
	}
}

// TestLoggerClose verifies that the logger can be closed properly
func TestLoggerClose(t *testing.T) {
	tempFile, err := os.CreateTemp("", "test_log_*.log")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	logger := NewLogger(tempFile, DEBUG)
	logger.LogFile = tempFile

	logger.Info("Test message")
	logger.Close()

	// Try to write to the closed file
	_, err = tempFile.Write([]byte("This should fail"))
	if err == nil {
		t.Error("Expected an error writing to closed file, but got none")
	}
}

// TestHandleLoggingError verifies that logging errors are handled correctly
func TestHandleLoggingError(t *testing.T) {
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	testError := fmt.Errorf("test error")
	handleLoggingError("TEST", testError)

	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	_, err := io.Copy(&buf, r)
	if err != nil {
		t.Fatalf("Failed to copy from pipe: %v", err)
	}
	output := buf.String()

	expectedOutput := "Logger TEST Output error: test error\n"
	if output != expectedOutput {
		t.Errorf("Expected output %q, but got %q", expectedOutput, output)
	}
}

// TestLogToFile verifies that logs are correctly written to a file
func TestLogToFile(t *testing.T) {
	tempFile, err := os.CreateTemp("", "test_log_*.log")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	cfg := &config.Config{
		LogLevel: "debug",
		LogFile:  tempFile.Name(),
	}

	logger := SetupLogger(cfg)
	defer logger.Close()

	testMessage := "Test log message"
	logger.Info(testMessage)

	content, err := os.ReadFile(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	if !bytes.Contains(content, []byte(testMessage)) {
		t.Errorf("Expected log file to contain %q, but it didn't", testMessage)
	}
}

// TestLogLevelFiltering verifies that log messages are filtered based on the set log level
func TestLogLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf, WARN)

	logger.Debug("This should not be logged")
	logger.Info("This should not be logged")
	logger.Warn("This should be logged")
	logger.Error("This should be logged")

	output := buf.String()
	if strings.Contains(output, "This should not be logged") {
		t.Error("Debug or Info message was logged when it shouldn't have been")
	}
	if !strings.Contains(output, "This should be logged") {
		t.Error("Warn or Error message was not logged when it should have been")
	}
}

// TestLogLevelString verifies that log levels are correctly converted to strings
func TestLogLevelString(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{DEBUG, "DEBUG"},
		{INFO, "INFO"},
		{WARN, "WARN"},
		{ERROR, "ERROR"},
		{LogLevel(99), "INFO"},
	}

	for _, tt := range tests {
		result := tt.level.String()
		if result != tt.expected {
			t.Errorf("LogLevel(%d).String() = %s, want %s", tt.level, result, tt.expected)
		}
	}
}
