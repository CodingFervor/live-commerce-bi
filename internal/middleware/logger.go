package middleware

import (
	"time"

	""+MOD+"/pkg/logger"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		c.Next()
		logger.Info("[HTTP] %s %s %d %s %s",
			c.Request.Method, path, c.Writer.Status(),
			time.Since(start).Round(time.Millisecond),
			c.ClientIP())
	}
}
