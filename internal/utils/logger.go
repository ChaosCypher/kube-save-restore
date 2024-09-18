package utils

import (
	"fmt"
	"io"
	"log"
	"os"

	"k8s-backup-restore/internal/config"
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
	level       LogLevel
	output      io.Writer
	debugLogger *log.Logger
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	errorLogger *log.Logger
	logFile     *os.File
}

// NewLogger creates a new Logger instance with the specified output and log level.
func NewLogger(out io.Writer, level LogLevel) *Logger {
	return &Logger{
		level:       level,
		output:      out,
		debugLogger: log.New(out, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile),
		infoLogger:  log.New(out, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		warnLogger:  log.New(out, "WARN: ", log.Ldate|log.Ltime|log.Lshortfile),
		errorLogger: log.New(out, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

// SetupLogger initializes the Logger based on the provided configuration.
// It sets up logging to a file if specified in the config, otherwise defaults to stdout.
func SetupLogger(config *config.Config) *Logger {
	var logWriter = os.Stdout
	var logFile *os.File
	if config.LogFile != "" {
		file, err := os.OpenFile(config.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			fmt.Printf("Failed to open log file: %v\n", err)
			os.Exit(1)
		}
		logWriter = file
		logFile = file
	}
	return &Logger{
		level:       parseLogLevel(config.LogLevel),
		output:      logWriter,
		debugLogger: log.New(logWriter, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile),
		infoLogger:  log.New(logWriter, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		warnLogger:  log.New(logWriter, "WARN: ", log.Ldate|log.Ltime|log.Lshortfile),
		errorLogger: log.New(logWriter, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
		logFile:     logFile,
	}
}

// Close closes the log file if it is open.
func (l *Logger) Close() {
	if l.logFile != nil {
		l.logFile.Close()
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
	if l.level <= DEBUG {
		l.debugLogger.Output(2, fmt.Sprintln(v...))
	}
}

// Info logs a message at the INFO level.
func (l *Logger) Info(v ...interface{}) {
	if l.level <= INFO {
		l.infoLogger.Output(2, fmt.Sprintln(v...))
	}
}

// Warn logs a message at the WARN level.
func (l *Logger) Warn(v ...interface{}) {
	if l.level <= WARN {
		l.warnLogger.Output(2, fmt.Sprintln(v...))
	}
}

// Error logs a message at the ERROR level.
func (l *Logger) Error(v ...interface{}) {
	if l.level <= ERROR {
		l.errorLogger.Output(2, fmt.Sprintln(v...))
	}
}

// Debugf logs a formatted message at the DEBUG level.
func (l *Logger) Debugf(format string, v ...interface{}) {
	if l.level <= DEBUG {
		l.debugLogger.Output(2, fmt.Sprintf(format, v...))
	}
}

// Infof logs a formatted message at the INFO level.
func (l *Logger) Infof(format string, v ...interface{}) {
	if l.level <= INFO {
		l.infoLogger.Output(2, fmt.Sprintf(format, v...))
	}
}

// Warnf logs a formatted message at the WARN level.
func (l *Logger) Warnf(format string, v ...interface{}) {
	if l.level <= WARN {
		l.warnLogger.Output(2, fmt.Sprintf(format, v...))
	}
}

// Errorf logs a formatted message at the ERROR level.
func (l *Logger) Errorf(format string, v ...interface{}) {
	if l.level <= ERROR {
		l.errorLogger.Output(2, fmt.Sprintf(format, v...))
	}
}
