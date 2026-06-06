package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/ranggaaprilio/keyles/infrastructure/config"
)

// SecurityHeaders sets security headers on all HTTP responses
func SecurityHeaders(cfg *config.Config) gin.HandlerFunc {
	csp := cfg.SecurityHeadersCSP
	if csp == "" {
		csp = "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'"
	}

	hsts := cfg.SecurityHeadersHSTS
	if hsts == "" {
		hsts = "max-age=31536000; includeSubDomains"
	}

	return func(c *gin.Context) {
		c.Header("Content-Security-Policy", csp)
		c.Header("Strict-Transport-Security", hsts)
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=(), payment=()")
		c.Header("Cross-Origin-Opener-Policy", "same-origin")
		c.Header("Cross-Origin-Embedder-Policy", "require-corp")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		c.Next()
	}
}
