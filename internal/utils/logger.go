package utils

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/chaoscypher/k8s-backup-restore/internal/config"
)

// LogLevel represents the severity of a log message.
type LogLevel int

// Log levels ordered by increasing severity.
const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

// Logger encapsulates logging functionality with support for different log levels.
type Logger struct {
	Level       LogLevel
	Output      io.Writer
	DebugLogger *log.Logger
	InfoLogger  *log.Logger
	WarnLogger  *log.Logger
	ErrorLogger *log.Logger
	LogFile     *os.File
}

// NewLogger creates a new Logger instance with the specified output and log level.
func NewLogger(out io.Writer, level LogLevel) *Logger {
	return &Logger{
		Level:       level,
		Output:      out,
		DebugLogger: log.New(out, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile),
		InfoLogger:  log.New(out, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		WarnLogger:  log.New(out, "WARN: ", log.Ldate|log.Ltime|log.Lshortfile),
		ErrorLogger: log.New(out, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

// SetupLogger initializes the Logger based on the provided configuration.
// It sets up logging to a file if specified in the config, otherwise defaults to stdout.
func SetupLogger(cfg *config.Config) *Logger {
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

// Close closes the log file if it is open.
func (l *Logger) Close() {
	if l.LogFile != nil {
		l.LogFile.Close()
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

// Debug logs a message at the DEBUG level.
func (l *Logger) Debug(v ...interface{}) {
	if l.Level <= DEBUG {
		err := l.DebugLogger.Output(2, fmt.Sprintln(v...))
		handleLoggingError("DEBUG", err)
	}
}

// Info logs a message at the INFO level.
func (l *Logger) Info(v ...interface{}) {
	if l.Level <= INFO {
		err := l.InfoLogger.Output(2, fmt.Sprintln(v...))
		handleLoggingError("INFO", err)
	}
}

// Warn logs a message at the WARN level.
func (l *Logger) Warn(v ...interface{}) {
	if l.Level <= WARN {
		err := l.WarnLogger.Output(2, fmt.Sprintln(v...))
		handleLoggingError("WARN", err)
	}
}

// Error logs a message at the ERROR level.
func (l *Logger) Error(v ...interface{}) {
	if l.Level <= ERROR {
		err := l.ErrorLogger.Output(2, fmt.Sprintln(v...))
		handleLoggingError("ERROR", err)
	}
}

// Debugf logs a formatted message at the DEBUG level.
func (l *Logger) Debugf(format string, v ...interface{}) {
	if l.Level <= DEBUG {
		err := l.DebugLogger.Output(2, fmt.Sprintf(format, v...))
		handleLoggingError("DEBUG", err)
	}
}

// Infof logs a formatted message at the INFO level.
func (l *Logger) Infof(format string, v ...interface{}) {
	if l.Level <= INFO {
		err := l.InfoLogger.Output(2, fmt.Sprintf(format, v...))
		handleLoggingError("INFO", err)
	}
}

// Warnf logs a formatted message at the WARN level.
func (l *Logger) Warnf(format string, v ...interface{}) {
	if l.Level <= WARN {
		err := l.WarnLogger.Output(2, fmt.Sprintf(format, v...))
		handleLoggingError("WARN", err)
	}
}

// Errorf logs a formatted message at the ERROR level.
func (l *Logger) Errorf(format string, v ...interface{}) {
	if l.Level <= ERROR {
		err := l.ErrorLogger.Output(2, fmt.Sprintf(format, v...))
		handleLoggingError("ERROR", err)
	}
}

// handleLoggingError handles errors returned from log.Logger.Output by writing them to stderr.
func handleLoggingError(prefix string, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Logger %s Output error: %v\n", prefix, err)
	}
}
