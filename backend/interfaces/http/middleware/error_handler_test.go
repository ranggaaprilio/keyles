package middleware

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeErrorEmailMasking(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "masks email in error",
			input:    "User user@example.com not found",
			expected: "User ***@*** not found",
		},
		{
			name:     "masks multiple emails",
			input:    "Contact admin@site.com or support@help.org",
			expected: "Contact ***@*** or ***@***",
		},
		{
			name:     "no email returns as is",
			input:    "Something went wrong",
			expected: "Something went wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeError(errors.New(tt.input))
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeErrorPathMasking(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "masks user path",
			input:    "File /Users/admin/data.txt not found",
			expected: "File [REDACTED] not found",
		},
		{
			name:     "masks home path",
			input:    "Config /home/user/app.yml missing",
			expected: "Config [REDACTED] missing",
		},
		{
			name:     "masks app path",
			input:    "Binary /app/server not executable",
			expected: "Binary [REDACTED] not executable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeError(errors.New(tt.input))
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeErrorSensitivePatterns(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "contains password",
			input: "Invalid password provided",
		},
		{
			name:  "contains secret",
			input: "Secret key is invalid",
		},
		{
			name:  "contains token",
			input: "Token expired",
		},
		{
			name:  "contains database",
			input: "Database connection failed",
		},
		{
			name:  "contains stack trace",
			input: "Stack trace: ...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeError(errors.New(tt.input))
			assert.Equal(t, "An error occurred. Please contact support if the problem persists.", result)
		})
	}
}

func TestSanitizeErrorNil(t *testing.T) {
	result := sanitizeError(nil)
	assert.Equal(t, "An error occurred", result)
}
