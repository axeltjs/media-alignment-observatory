package middleware

import (
	"github.com/gin-gonic/gin"
)

func CORS(allowedOrigins string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" || allowedOrigins == "*" {
			origin = "*"
		}
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
