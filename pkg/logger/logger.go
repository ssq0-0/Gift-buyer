package logger

import (
	"io"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/writer"
)

// Package logger provides a logging interface and implementation using logrus.
// It supports different log levels, structured logging with fields, and configurable output.
// The package offers both a global logger instance and factory methods for creating
// custom logger instances with specific configurations.

// GlobalLogger is a singleton instance of the logger that can be used throughout the application.
// It provides a convenient way to access logging functionality without explicitly creating
// logger instances in every package.
var GlobalLogger Logger

// init initializes the global logger with InfoLevel as the default logging level.
func init() {
	GlobalLogger = New(InfoLevel)
}

// Init reinitializes the global logger with the specified logging level.
// This function should be called early in the application lifecycle to configure
// the desired logging level for the entire application.
//
// Parameters:
//   - level: the desired logging level for the global logger
func Init(level LoggerLevel) {
	GlobalLogger = New(level)
}

// ParseLevel converts a string representation of a logging level to LoggerLevel.
// It supports common level names and provides case-insensitive matching.
// If an unknown level is provided, it defaults to InfoLevel.
//
// Supported levels:
//   - "debug" -> DebugLevel
//   - "info" -> InfoLevel
//   - "warn", "warning" -> WarnLevel
//   - "error" -> ErrorLevel
//   - "fatal" -> FatalLevel
//   - "panic" -> PanicLevel
//
// Parameters:
//   - levelStr: string representation of the logging level
//
// Returns:
//   - LoggerLevel: parsed logging level, defaults to InfoLevel for unknown values
func ParseLevel(levelStr string) LoggerLevel {
	levelStr = strings.ToLower(levelStr)
	switch levelStr {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warn", "warning":
		return WarnLevel
	case "error":
		return ErrorLevel
	case "fatal":
		return FatalLevel
	case "panic":
		return PanicLevel
	default:
		return InfoLevel
	}
}

// Fields represents a map of key-value pairs that can be added to log entries.
// It provides structured logging capabilities by allowing additional context
// to be attached to log messages.
type Fields map[string]interface{}

// LoggerLevel represents the severity level of log messages.
// It defines the importance and urgency of log entries, allowing for
// filtering and routing based on message severity.
type LoggerLevel string

const (
	// DebugLevel represents debug level logging for detailed diagnostic information.
	// Typically used during development and troubleshooting.
	DebugLevel LoggerLevel = "debug"

	// InfoLevel represents informational level logging for general application flow.
	// Used for tracking normal application behavior and important events.
	InfoLevel LoggerLevel = "info"

	// WarnLevel represents warning level logging for potentially harmful situations.
	// Indicates issues that don't prevent operation but should be investigated.
	WarnLevel LoggerLevel = "warn"

	// ErrorLevel represents error level logging for error events that don't stop execution.
	// Used for recoverable errors and exceptional conditions.
	ErrorLevel LoggerLevel = "error"

	// FatalLevel represents fatal level logging for severe errors that cause program termination.
	// The application will exit after logging a fatal message.
	FatalLevel LoggerLevel = "fatal"

	// PanicLevel represents panic level logging for severe errors that cause a panic.
	// The application will panic after logging a panic message.
	PanicLevel LoggerLevel = "panic"
)

// Logger defines the interface for logging operations.
// It provides methods for logging at different severity levels
// and supports structured logging with fields for enhanced context.
type Logger interface {
	// Debug logs a message at debug level.
	// Used for detailed diagnostic information typically only of interest
	// when diagnosing problems.
	Debug(args ...interface{})

	// Debugf logs a formatted message at debug level.
	// Supports printf-style formatting for dynamic message construction.
	Debugf(format string, args ...interface{})

	// Error logs a message at error level.
	// Used for error events that might still allow the application to continue running.
	Error(args ...interface{})

	// Errorf logs a formatted message at error level.
	// Supports printf-style formatting for dynamic error message construction.
	Errorf(format string, args ...interface{})

	// Fatal logs a message at fatal level and exits the program.
	// This should be used for errors that prevent the application from continuing.
	Fatal(args ...interface{})

	// Fatalf logs a formatted message at fatal level and exits the program.
	// Supports printf-style formatting for dynamic fatal message construction.
	Fatalf(format string, args ...interface{})

	// Info logs a message at info level.
	// Used for general informational messages that highlight application progress.
	Info(args ...interface{})

	// Infof logs a formatted message at info level.
	// Supports printf-style formatting for dynamic informational messages.
	Infof(format string, args ...interface{})

	// Panic logs a message at panic level and panics.
	// This should be used for errors that represent a programming error.
	Panic(args ...interface{})

	// Panicf logs a formatted message at panic level and panics.
	// Supports printf-style formatting for dynamic panic messages.
	Panicf(format string, args ...interface{})

	// Warn logs a message at warn level.
	// Used for potentially harmful situations that don't prevent operation.
	Warn(args ...interface{})

	// Warnf logs a formatted message at warn level.
	// Supports printf-style formatting for dynamic warning messages.
	Warnf(format string, args ...interface{})

	// WithFields returns a new logger instance with the specified fields.
	// The fields will be included in all subsequent log messages from the returned logger.
	WithFields(fields Fields) Logger
}

// logger implements the Logger interface using logrus as the underlying logging framework.
// It provides thread-safe logging with configurable output and formatting options.
type logger struct {
	*logrus.Logger
	entry *logrus.Entry
}

// New creates a new logger instance with the specified log level.
// The logger is configured with colored text output, timestamps, and
// asynchronous writing to stdout for optimal performance.
//
// Configuration details:
//   - Uses TextFormatter with colors and full timestamps
//   - Outputs to stdout via asynchronous writer hook
//   - Supports all standard logrus logging levels
//   - Thread-safe for concurrent use
//
// Parameters:
//   - level: the logging level for filtering messages
//
// Returns:
//   - Logger: configured logger instance ready for use
func New(level LoggerLevel) Logger {
	logrusLevel, err := logrus.ParseLevel(string(level))
	if err != nil {
		logrusLevel = logrus.InfoLevel
	}

	lgr := logrus.New()
	lgr.SetLevel(logrusLevel)
	lgr.SetOutput(io.Discard)

	// Configure text formatter with colors and timestamps
	lgr.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})

	// Add asynchronous hook for writing to stdout
	lgr.AddHook(&writer.Hook{
		Writer:    os.Stdout,
		LogLevels: logrus.AllLevels,
	})

	return logger{Logger: lgr}
}

// WithFields returns a new logger instance with the specified fields.
// The fields will be included in all subsequent log messages, providing
// structured logging capabilities for enhanced context and filtering.
//
// Parameters:
//   - fields: map of key-value pairs to include in log messages
//
// Returns:
//   - Logger: new logger instance with attached fields
func (l logger) WithFields(fields Fields) Logger {
	if l.entry != nil {
		return logger{
			Logger: l.Logger,
			entry:  l.entry.WithFields(logrus.Fields(fields)),
		}
	}
	return logger{
		Logger: l.Logger,
		entry:  l.Logger.WithFields(logrus.Fields(fields)),
	}
}

// getEntry returns the appropriate logrus entry or logger for message output.
// It handles the internal routing between field-enhanced entries and the base logger.
//
// Returns:
//   - logrus.FieldLogger: the appropriate logger instance for message output
func (l logger) getEntry() logrus.FieldLogger {
	if l.entry != nil {
		return l.entry
	}
	return l.Logger
}

// Debug logs a message at debug level.
func (l logger) Debug(args ...interface{}) {
	l.getEntry().Debug(args...)
}

// Debugf logs a formatted message at debug level.
func (l logger) Debugf(format string, args ...interface{}) {
	l.getEntry().Debugf(format, args...)
}

// Error logs a message at error level.
func (l logger) Error(args ...interface{}) {
	l.getEntry().Error(args...)
}

// Errorf logs a formatted message at error level.
func (l logger) Errorf(format string, args ...interface{}) {
	l.getEntry().Errorf(format, args...)
}

// Fatal logs a message at fatal level and exits the program.
func (l logger) Fatal(args ...interface{}) {
	l.getEntry().Fatal(args...)
}

// Fatalf logs a formatted message at fatal level and exits the program.
func (l logger) Fatalf(format string, args ...interface{}) {
	l.getEntry().Fatalf(format, args...)
}

// Info logs a message at info level.
func (l logger) Info(args ...interface{}) {
	l.getEntry().Info(args...)
}

// Infof logs a formatted message at info level.
func (l logger) Infof(format string, args ...interface{}) {
	l.getEntry().Infof(format, args...)
}

// Panic logs a message at panic level and panics.
func (l logger) Panic(args ...interface{}) {
	l.getEntry().Panic(args...)
}

// Panicf logs a formatted message at panic level and panics.
func (l logger) Panicf(format string, args ...interface{}) {
	l.getEntry().Panicf(format, args...)
}

// Warn logs a message at warn level.
func (l logger) Warn(args ...interface{}) {
	l.getEntry().Warn(args...)
}

// Warnf logs a formatted message at warn level.
func (l logger) Warnf(format string, args ...interface{}) {
	l.getEntry().Warnf(format, args...)
}
