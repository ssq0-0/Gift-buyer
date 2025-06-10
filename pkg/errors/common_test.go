package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected string
	}{
		{
			name:     "creates error with simple text",
			text:     "test error",
			expected: "test error",
		},
		{
			name:     "creates error with empty text",
			text:     "",
			expected: "",
		},
		{
			name:     "creates error with special characters",
			text:     "error: failed to process $data with 100% accuracy",
			expected: "error: failed to process $data with 100% accuracy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := New(tt.text)
			assert.Error(t, err)
			assert.Equal(t, tt.expected, err.Error())
		})
	}
}

func TestWrap(t *testing.T) {
	originalErr := errors.New("original error")

	tests := []struct {
		name     string
		err      error
		context  string
		expected string
		isNil    bool
	}{
		{
			name:     "wraps error with context",
			err:      originalErr,
			context:  "operation failed",
			expected: "operation failed: original error",
			isNil:    false,
		},
		{
			name:     "returns nil when wrapping nil error",
			err:      nil,
			context:  "some context",
			expected: "",
			isNil:    true,
		},
		{
			name:     "wraps error with empty context",
			err:      originalErr,
			context:  "",
			expected: ": original error",
			isNil:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Wrap(tt.err, tt.context)
			if tt.isNil {
				assert.Nil(t, result)
			} else {
				assert.Error(t, result)
				assert.Equal(t, tt.expected, result.Error())
				assert.True(t, errors.Is(result, tt.err))
			}
		})
	}
}

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		text string
	}{
		{"ErrNotFound", ErrNotFound, "not found"},
		{"ErrUnsupportedType", ErrUnsupportedType, "unsupported type"},
		{"ErrExitProgram", ErrExitProgram, "exit program"},
		{"ErrValueEmpty", ErrValueEmpty, "value empty"},
		{"ErrSelection", ErrSelection, "selection error"},
		{"ErrUnexpectedType", ErrUnexpectedType, "unexpected type"},
		{"ErrConnectionFailed", ErrConnectionFailed, "connection failed"},
		{"ErrRequestFailed", ErrRequestFailed, "request failed"},
		{"ErrResponseParsing", ErrResponseParsing, "failed to parse response"},
		{"ErrStatusCode", ErrStatusCode, "unexpected status code"},
		{"ErrRateLimitReached", ErrRateLimitReached, "rate limit reached"},
		{"ErrInvalidParams", ErrInvalidParams, "invalid or missing parameters"},
		{"ErrNoCreatedValue", ErrNoCreatedValue, "no parameter has been created"},
		{"ErrFailedInit", ErrFailedInit, "failed to initialize"},
		{"ErrConfigRead", ErrConfigRead, "failed to read config"},
		{"ErrConfigParse", ErrConfigParse, "failed to parse config"},
		{"ErrConfigSave", ErrConfigSave, "failed to save config"},
		{"ErrInvalidConfig", ErrInvalidConfig, "invalid configuration"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Error(t, tt.err)
			assert.Equal(t, tt.text, tt.err.Error())
		})
	}
}

func TestErrorWrapping(t *testing.T) {
	baseErr := ErrNotFound
	wrappedErr := Wrap(baseErr, "user lookup failed")

	assert.Error(t, wrappedErr)
	assert.Equal(t, "user lookup failed: not found", wrappedErr.Error())
	assert.True(t, errors.Is(wrappedErr, baseErr))
}

func TestMultipleWrapping(t *testing.T) {
	baseErr := ErrConnectionFailed
	firstWrap := Wrap(baseErr, "database connection")
	secondWrap := Wrap(firstWrap, "service initialization")

	assert.Error(t, secondWrap)
	assert.Equal(t, "service initialization: database connection: connection failed", secondWrap.Error())
	assert.True(t, errors.Is(secondWrap, baseErr))
	assert.True(t, errors.Is(secondWrap, firstWrap))
}
