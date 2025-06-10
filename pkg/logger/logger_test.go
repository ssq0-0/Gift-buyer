package logger

import (
	"bytes"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestParseLevel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected LoggerLevel
	}{
		{"debug level", "debug", DebugLevel},
		{"info level", "info", InfoLevel},
		{"warn level", "warn", WarnLevel},
		{"warning level", "warning", WarnLevel},
		{"error level", "error", ErrorLevel},
		{"fatal level", "fatal", FatalLevel},
		{"panic level", "panic", PanicLevel},
		{"uppercase debug", "DEBUG", DebugLevel},
		{"mixed case info", "InFo", InfoLevel},
		{"invalid level defaults to info", "invalid", InfoLevel},
		{"empty string defaults to info", "", InfoLevel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseLevel(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name  string
		level LoggerLevel
	}{
		{"debug logger", DebugLevel},
		{"info logger", InfoLevel},
		{"warn logger", WarnLevel},
		{"error logger", ErrorLevel},
		{"fatal logger", FatalLevel},
		{"panic logger", PanicLevel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := New(tt.level)
			assert.NotNil(t, logger)
			assert.Implements(t, (*Logger)(nil), logger)
		})
	}
}

func TestLoggerMethods(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer

	// Create a custom logrus logger for testing
	logrusLogger := logrus.New()
	logrusLogger.SetOutput(&buf)
	logrusLogger.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp: true,
		DisableColors:    true,
	})
	logrusLogger.SetLevel(logrus.DebugLevel)

	logger := &logger{Logger: logrusLogger}

	tests := []struct {
		name     string
		logFunc  func()
		expected string
		level    string
	}{
		{
			name: "debug message",
			logFunc: func() {
				buf.Reset()
				logger.Debug("debug message")
			},
			expected: "debug message",
			level:    "debug",
		},
		{
			name: "info message",
			logFunc: func() {
				buf.Reset()
				logger.Info("info message")
			},
			expected: "info message",
			level:    "info",
		},
		{
			name: "warn message",
			logFunc: func() {
				buf.Reset()
				logger.Warn("warn message")
			},
			expected: "warn message",
			level:    "warning",
		},
		{
			name: "error message",
			logFunc: func() {
				buf.Reset()
				logger.Error("error message")
			},
			expected: "error message",
			level:    "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.logFunc()
			output := buf.String()
			assert.Contains(t, strings.ToLower(output), tt.level)
			assert.Contains(t, output, tt.expected)
		})
	}
}

func TestLoggerFormattedMethods(t *testing.T) {
	var buf bytes.Buffer

	logrusLogger := logrus.New()
	logrusLogger.SetOutput(&buf)
	logrusLogger.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp: true,
		DisableColors:    true,
	})
	logrusLogger.SetLevel(logrus.DebugLevel)

	logger := &logger{Logger: logrusLogger}

	tests := []struct {
		name     string
		logFunc  func()
		expected string
		level    string
	}{
		{
			name: "debugf message",
			logFunc: func() {
				buf.Reset()
				logger.Debugf("debug %s %d", "test", 123)
			},
			expected: "debug test 123",
			level:    "debug",
		},
		{
			name: "infof message",
			logFunc: func() {
				buf.Reset()
				logger.Infof("info %s %d", "test", 456)
			},
			expected: "info test 456",
			level:    "info",
		},
		{
			name: "warnf message",
			logFunc: func() {
				buf.Reset()
				logger.Warnf("warn %s %d", "test", 789)
			},
			expected: "warn test 789",
			level:    "warning",
		},
		{
			name: "errorf message",
			logFunc: func() {
				buf.Reset()
				logger.Errorf("error %s %d", "test", 999)
			},
			expected: "error test 999",
			level:    "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.logFunc()
			output := buf.String()
			assert.Contains(t, strings.ToLower(output), tt.level)
			assert.Contains(t, output, tt.expected)
		})
	}
}

func TestWithFields(t *testing.T) {
	var buf bytes.Buffer

	logrusLogger := logrus.New()
	logrusLogger.SetOutput(&buf)
	logrusLogger.SetFormatter(&logrus.JSONFormatter{})
	logrusLogger.SetLevel(logrus.InfoLevel)

	logger := &logger{Logger: logrusLogger}

	fields := Fields{
		"user_id": 123,
		"action":  "login",
		"ip":      "192.168.1.1",
	}

	fieldLogger := logger.WithFields(fields)
	assert.NotNil(t, fieldLogger)
	assert.Implements(t, (*Logger)(nil), fieldLogger)

	fieldLogger.Info("user logged in")
	output := buf.String()

	// Check that all fields are present in the JSON output
	assert.Contains(t, output, "user_id")
	assert.Contains(t, output, "123")
	assert.Contains(t, output, "action")
	assert.Contains(t, output, "login")
	assert.Contains(t, output, "ip")
	assert.Contains(t, output, "192.168.1.1")
	assert.Contains(t, output, "user logged in")
}

func TestLogLevels(t *testing.T) {
	tests := []struct {
		name        string
		loggerLevel LoggerLevel
		logLevel    string
		shouldLog   bool
	}{
		{"debug logger logs debug", DebugLevel, "debug", true},
		{"debug logger logs info", DebugLevel, "info", true},
		{"debug logger logs warn", DebugLevel, "warn", true},
		{"debug logger logs error", DebugLevel, "error", true},
		{"info logger skips debug", InfoLevel, "debug", false},
		{"info logger logs info", InfoLevel, "info", true},
		{"info logger logs warn", InfoLevel, "warn", true},
		{"info logger logs error", InfoLevel, "error", true},
		{"warn logger skips debug", WarnLevel, "debug", false},
		{"warn logger skips info", WarnLevel, "info", false},
		{"warn logger logs warn", WarnLevel, "warn", true},
		{"warn logger logs error", WarnLevel, "error", true},
		{"error logger skips debug", ErrorLevel, "debug", false},
		{"error logger skips info", ErrorLevel, "info", false},
		{"error logger skips warn", ErrorLevel, "warn", false},
		{"error logger logs error", ErrorLevel, "error", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			logrusLogger := logrus.New()
			logrusLogger.SetOutput(&buf)
			logrusLogger.SetFormatter(&logrus.TextFormatter{
				DisableTimestamp: true,
				DisableColors:    true,
			})

			// Set the logger level
			logrusLevel, _ := logrus.ParseLevel(string(tt.loggerLevel))
			logrusLogger.SetLevel(logrusLevel)

			logger := &logger{Logger: logrusLogger}

			// Log at the specified level
			switch tt.logLevel {
			case "debug":
				logger.Debug("test message")
			case "info":
				logger.Info("test message")
			case "warn":
				logger.Warn("test message")
			case "error":
				logger.Error("test message")
			}

			output := buf.String()
			if tt.shouldLog {
				assert.Contains(t, output, "test message")
			} else {
				assert.Empty(t, output)
			}
		})
	}
}

func TestMultipleArguments(t *testing.T) {
	var buf bytes.Buffer

	logrusLogger := logrus.New()
	logrusLogger.SetOutput(&buf)
	logrusLogger.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp: true,
		DisableColors:    true,
	})
	logrusLogger.SetLevel(logrus.InfoLevel)

	logger := &logger{Logger: logrusLogger}

	logger.Info("multiple", "arguments", 123, true)
	output := buf.String()

	assert.Contains(t, output, "multiple")
	assert.Contains(t, output, "arguments")
	assert.Contains(t, output, "123")
	assert.Contains(t, output, "true")
}
