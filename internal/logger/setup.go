package logger

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/chaoscypher/kube-save-restore/internal/config"
)

// SetupLogger initializes the Logger based on the provided configuration
func SetupLogger(cfg *config.Config) LoggerInterface {
	var logWriter io.Writer = os.Stdout
	var logFile *os.File

	if cfg.LogFile != "" {
		file, err := os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			fmt.Printf("Failed to open log file: %v\n", err)
			os.Exit(1)
		}
		logWriter = file
		logFile = file
	}

	logger := NewLogger(logWriter, parseLogLevel(cfg.LogLevel))
	logger.LogFile = logFile

	return logger
}

// parseLogLevel converts a string log level to the corresponding LogLevel type
func parseLogLevel(level string) LogLevel {
	switch strings.ToLower(level) {
	case "debug":
		return DEBUG
	case "warn", "warning":
		return WARN
	case "error":
		return ERROR
	default:
		return INFO
	}
}
