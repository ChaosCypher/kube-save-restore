package logger

import (
	"fmt"
	"io"
	"log"
	"os"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

// String provides the string representation of the LogLevel
func (ll LogLevel) String() string {
	switch ll {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "INFO"
	}
}

// LoggerInterface defines the methods for logging
type LoggerInterface interface {
	Debug(v ...interface{})
	Info(v ...interface{})
	Warn(v ...interface{})
	Error(v ...interface{})
	Debugf(format string, v ...interface{})
	Infof(format string, v ...interface{})
	Warnf(format string, v ...interface{})
	Errorf(format string, v ...interface{})
	Close()
}

// Logger encapsulates logging functionality with support for different log levels
type Logger struct {
	Level   LogLevel
	Output  io.Writer
	Logger  *log.Logger
	LogFile *os.File
}

// NewLogger creates a new Logger instance with the specified output and log level
func NewLogger(out io.Writer, level LogLevel) *Logger {
	return &Logger{
		Level:  level,
		Output: out,
		Logger: log.New(out, "", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

// Close closes the log file if it is open
func (l *Logger) Close() {
	if l.LogFile != nil {
		l.LogFile.Close()
	}
}

// logMessage logs a message with the given level
func (l *Logger) logMessage(level LogLevel, msg string) {
	if level >= l.Level {
		prefix := fmt.Sprintf("%s: ", level.String())
		err := l.Logger.Output(3, prefix+msg)
		handleLoggingError(level.String(), err)
	}
}

// log logs a message at the given level
func (l *Logger) log(level LogLevel, v ...interface{}) {
	l.logMessage(level, fmt.Sprint(v...))
}

// logf logs a formatted message at the given level
func (l *Logger) logf(level LogLevel, format string, v ...interface{}) {
	l.logMessage(level, fmt.Sprintf(format, v...))
}

// Debug logs a message at the DEBUG level
func (l *Logger) Debug(v ...interface{}) { l.log(DEBUG, v...) }

// Info logs a message at the INFO level
func (l *Logger) Info(v ...interface{}) { l.log(INFO, v...) }

// Warn logs a message at the WARN level
func (l *Logger) Warn(v ...interface{}) { l.log(WARN, v...) }

// Error logs a message at the ERROR level
func (l *Logger) Error(v ...interface{}) { l.log(ERROR, v...) }

// Debugf logs a formatted message at the DEBUG level
func (l *Logger) Debugf(format string, v ...interface{}) { l.logf(DEBUG, format, v...) }

// Infof logs a formatted message at the INFO level
func (l *Logger) Infof(format string, v ...interface{}) { l.logf(INFO, format, v...) }

// Warnf logs a formatted message at the WARN level
func (l *Logger) Warnf(format string, v ...interface{}) { l.logf(WARN, format, v...) }

// Errorf logs a formatted message at the ERROR level
func (l *Logger) Errorf(format string, v ...interface{}) { l.logf(ERROR, format, v...) }

// handleLoggingError handles errors returned from log.Logger.Output by writing them to stderr
func handleLoggingError(prefix string, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Logger %s Output error: %v\n", prefix, err)
	}
}
