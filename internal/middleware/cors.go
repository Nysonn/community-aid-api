package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORS returns a middleware that sets CORS headers.
// allowedOrigins is a comma-separated list of permitted origins.
// If empty, all origins are allowed (development fallback).
func CORS(allowedOrigins string) gin.HandlerFunc {
	allowAll := strings.TrimSpace(allowedOrigins) == ""

	originSet := make(map[string]bool)
	if !allowAll {
		for _, o := range strings.Split(allowedOrigins, ",") {
			if trimmed := strings.TrimSpace(o); trimmed != "" {
				originSet[trimmed] = true
			}
		}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		if allowAll {
			c.Header("Access-Control-Allow-Origin", "*")
		} else if originSet[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Authorization, Content-Type, Accept")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "43200")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
