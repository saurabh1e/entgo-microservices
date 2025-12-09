package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORS middleware for handling Cross-Origin Resource Sharing
func CORS() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		requestHeaders := c.Request.Header.Get("Access-Control-Request-Headers")
		requestMethod := c.Request.Header.Get("Access-Control-Request-Method")
		requestPrivateNetwork := c.Request.Header.Get("Access-Control-Request-Private-Network")

		// Allow all origins (no credentials). If you need credentials, echo the origin instead of '*'.
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "false")
		c.Header("Access-Control-Allow-Headers", coalesce(requestHeaders, "Authorization, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, accept, origin, Cache-Control, X-Requested-With"))
		c.Header("Access-Control-Allow-Methods", coalesce(requestMethod, "GET, POST, OPTIONS, PUT, DELETE"))
		c.Header("Access-Control-Max-Age", "600")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Content-Type")

		// Properly set Vary header
		if requestMethod != "" || requestHeaders != "" || requestPrivateNetwork != "" {
			c.Header("Vary", "Origin, Access-Control-Request-Method, Access-Control-Request-Headers, Access-Control-Request-Private-Network")
		} else {
			c.Header("Vary", "Origin")
		}

		// Handle Private Network preflight (Chrome feature) when requested
		if strings.EqualFold(requestPrivateNetwork, "true") || strings.EqualFold(requestPrivateNetwork, "?1") {
			c.Header("Access-Control-Allow-Private-Network", "true")
		}

		if c.Request.Method == http.MethodOptions {
			// For preflight, return 204 with the above headers
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		// Proceed with request
		_ = origin // origin not used when allowing '*', kept for potential future use
		c.Next()
	})
}

func coalesce(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// Rate limit header constants
const (
	HeaderRateLimitLimit     = "X-RateLimit-Limit"
	HeaderRateLimitRemaining = "X-RateLimit-Remaining"
	HeaderRateLimitReset     = "X-RateLimit-Reset"
	HeaderRetryAfter         = "Retry-After"
)
