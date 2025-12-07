package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// ErrorHandler is a middleware that handles errors and sanitizes error messages
// per FR-021: provide clear error messages without exposing security details
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any errors
		if len(c.Errors) == 0 {
			return
		}

		// Get the last error
		err := c.Errors.Last()
		
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

// RecoveryHandler handles panics and prevents stack trace exposure
func RecoveryHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the full error internally (would integrate with logging system)
				// But only return sanitized error to client
				
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
