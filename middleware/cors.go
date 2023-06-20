package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func CorsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Authorization, Content-Type, Cache-Control, expires")
		c.Writer.Header().Set("Access-Control-Max-Age", "3600")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		// Check if the request is a multipart request
		contentType := c.Request.Header.Get("Content-Type")
		if strings.HasPrefix(contentType, "multipart/form-data") {
			// Add the "Content-Type" header for multipart response
			c.Writer.Header().Set("Content-Type", "multipart/form-data")
		}

		c.Next()
	}
}
