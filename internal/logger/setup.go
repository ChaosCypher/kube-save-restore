package logger

import (
	"fmt"
	"log"
	"os"

	"github.com/chaoscypher/k8s-backup-restore/internal/config"
)

// SetupLogger initializes the Logger based on the provided configuration.
// It sets up logging to a file if specified in the config, otherwise defaults to stdout.
func SetupLogger(cfg *config.Config) LoggerInterface {
	var logWriter = os.Stdout
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
	return &Logger{
		Level:       parseLogLevel(cfg.LogLevel),
		Output:      logWriter,
		DebugLogger: log.New(logWriter, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile),
		InfoLogger:  log.New(logWriter, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		WarnLogger:  log.New(logWriter, "WARN: ", log.Ldate|log.Ltime|log.Lshortfile),
		ErrorLogger: log.New(logWriter, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
		LogFile:     logFile,
	}
}

// parseLogLevel converts a string log level to the corresponding LogLevel type.
func parseLogLevel(level string) LogLevel {
	switch level {
	case "debug":
		return DEBUG
	case "warn":
		return WARN
	case "error":
		return ERROR
	default:
		return INFO
	}
}
