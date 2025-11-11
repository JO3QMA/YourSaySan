package logger

import (
	"log"
	"os"
)

// Logger represents a simple logger interface
type Logger interface {
	Info(format string, args ...interface{})
	Error(format string, args ...interface{})
	Debug(format string, args ...interface{})
}

// DefaultLogger implements Logger using standard log package
type DefaultLogger struct {
	infoLogger  *log.Logger
	errorLogger *log.Logger
	debugLogger *log.Logger
	debug       bool
}

// NewDefaultLogger creates a new default logger
func NewDefaultLogger(debug bool) Logger {
	return &DefaultLogger{
		infoLogger:  log.New(os.Stdout, "[INFO] ", log.LstdFlags),
		errorLogger: log.New(os.Stderr, "[ERROR] ", log.LstdFlags),
		debugLogger: log.New(os.Stdout, "[DEBUG] ", log.LstdFlags),
		debug:       debug,
	}
}

// Info logs info level messages
func (l *DefaultLogger) Info(format string, args ...interface{}) {
	l.infoLogger.Printf(format, args...)
}

// Error logs error level messages
func (l *DefaultLogger) Error(format string, args ...interface{}) {
	l.errorLogger.Printf(format, args...)
}

// Debug logs debug level messages (only if debug is enabled)
func (l *DefaultLogger) Debug(format string, args ...interface{}) {
	if l.debug {
		l.debugLogger.Printf(format, args...)
	}
}

// NoOpLogger is a logger that does nothing
type NoOpLogger struct{}

// NewNoOpLogger creates a no-op logger
func NewNoOpLogger() Logger {
	return &NoOpLogger{}
}

// Info does nothing
func (l *NoOpLogger) Info(format string, args ...interface{}) {}

// Error does nothing
func (l *NoOpLogger) Error(format string, args ...interface{}) {}

// Debug does nothing
func (l *NoOpLogger) Debug(format string, args ...interface{}) {}
