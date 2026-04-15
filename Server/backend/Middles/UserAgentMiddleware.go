package MiddleWare

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func UserAgentCheckMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userAgent := c.Request.UserAgent()

		if strings.Contains(strings.ToLower(userAgent), "aia gent") || 
		   strings.Contains(strings.ToLower(userAgent), "ai-agent") ||
		   strings.Contains(strings.ToLower(userAgent), "aia gent") {
			c.JSON(http.StatusForbidden, gin.H{
				"code":  http.StatusForbidden,
				"error": "AI Agent access not allowed",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
