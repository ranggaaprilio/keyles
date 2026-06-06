package middleware

import (
	"fmt"
	"net/http"
	"regexp"
	"runtime/debug"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ory/fosite"
	"github.com/ranggaaprilio/keyles/infrastructure/logging"
	"github.com/ranggaaprilio/keyles/infrastructure/monitoring"
)

// Logger is the structured logger used by middleware
var Logger *logging.Logger

func init() {
	Logger = logging.NewLogger("info")
}

// ErrorHandler is a middleware that handles errors and sanitizes error messages
// per FR-021: provide clear error messages without exposing security details
// Updated to handle OAuth 2.0 specific errors per FR-044
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any errors
		if len(c.Errors) == 0 {
			return
		}

		// Get the last error
		err := c.Errors.Last()

		// Check if it's an OAuth error
		if oauthErr, ok := err.Err.(*fosite.RFC6749Error); ok {
			handleOAuthError(c, oauthErr)
			return
		}

		// Sanitize error message
		sanitized := sanitizeError(err.Err)

		// Determine status code
		status := c.Writer.Status()
		if status == http.StatusOK {
			status = http.StatusInternalServerError
		}

		c.JSON(status, gin.H{
			"success": false,
			"error": gin.H{
				"code":    getErrorCode(status),
				"message": sanitized,
			},
		})
	}
}

// handleOAuthError handles OAuth 2.0 specific errors per RFC 6749
func handleOAuthError(c *gin.Context, err *fosite.RFC6749Error) {
	status := http.StatusBadRequest

	// Map OAuth error codes to HTTP status codes
	switch err.ErrorField {
	case "unauthorized_client", "access_denied":
		status = http.StatusForbidden
	case "invalid_client":
		status = http.StatusUnauthorized
	case "server_error":
		status = http.StatusInternalServerError
	case "temporarily_unavailable":
		status = http.StatusServiceUnavailable
	}

	// Return OAuth 2.0 formatted error
	response := gin.H{
		"error":             err.ErrorField,
		"error_description": err.DescriptionField,
	}

	if err.HintField != "" {
		response["error_hint"] = err.HintField
	}

	c.JSON(status, response)
}

// RecoveryHandler handles panics and prevents stack trace exposure
func RecoveryHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
			// Log the full error and stack trace internally
			Logger.Error("panic recovered",
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"error", err,
				"stack", string(debug.Stack()),
			)
			monitoring.IncrementSecurityEvent("tls_error")

				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error": gin.H{
						"code":    "INTERNAL_SERVER_ERROR",
						"message": "An unexpected error occurred. Please try again later.",
					},
				})

				c.Abort()
			}
		}()

		c.Next()
	}
}

// sanitizeError removes sensitive information from error messages
func sanitizeError(err error) string {
	if err == nil {
		return "An error occurred"
	}

	errMsg := err.Error()

	// Mask email addresses
	emailRegex := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	errMsg = emailRegex.ReplaceAllString(errMsg, "***@***")

	// Mask internal file paths
	pathPatterns := []string{"/Users/", "/home/", "/app/", "/var/", "/opt/", "/tmp/"}
	for _, path := range pathPatterns {
		if strings.Contains(errMsg, path) {
			// Replace the path and the following word (filename) with [REDACTED]
			pathRegex := regexp.MustCompile(regexp.QuoteMeta(path) + `[^\s:,;]+`)
			errMsg = pathRegex.ReplaceAllString(errMsg, "[REDACTED]")
		}
	}

	// List of sensitive patterns to hide
	sensitivePatterns := []string{
		"password",
		"secret",
		"token",
		"key",
		"database",
		"connection",
		"postgres",
		"redis",
		"stack trace",
		"panic",
		"runtime error",
	}

	// Check if error contains sensitive information
	lowerMsg := strings.ToLower(errMsg)
	for _, pattern := range sensitivePatterns {
		if strings.Contains(lowerMsg, pattern) {
			return "An error occurred. Please contact support if the problem persists."
		}
	}

	// Return the error message but limit length
	if len(errMsg) > 200 {
		return errMsg[:200] + "..."
	}

	return errMsg
}

// getErrorCode returns a user-friendly error code based on status
func getErrorCode(status int) string {
	codes := map[int]string{
		http.StatusBadRequest:          "BAD_REQUEST",
		http.StatusUnauthorized:        "UNAUTHORIZED",
		http.StatusForbidden:           "FORBIDDEN",
		http.StatusNotFound:            "NOT_FOUND",
		http.StatusConflict:            "CONFLICT",
		http.StatusTooManyRequests:     "RATE_LIMIT_EXCEEDED",
		http.StatusInternalServerError: "INTERNAL_SERVER_ERROR",
		http.StatusServiceUnavailable:  "SERVICE_UNAVAILABLE",
	}

	if code, ok := codes[status]; ok {
		return code
	}

	return fmt.Sprintf("ERROR_%d", status)
}
