package middleware

import (
	"bytes"
	"io"
	"time"

	"github.com/CodingFervor/live-commerce-bi/internal/model"
	"github.com/CodingFervor/live-commerce-bi/internal/repository"
	"github.com/CodingFervor/live-commerce-bi/pkg/logger"

	"github.com/gin-gonic/gin"
)

// AuditLogger records all API operations to audit_logs table
func AuditLogger() gin.HandlerFunc {
	auditRepo := repository.NewAuditRepo()

	return func(c *gin.Context) {
		start := time.Now()

		// Read request body for detail capture
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		c.Next()

		duration := time.Since(start).Milliseconds()

		userID, _ := c.Get("user_id")
		username, _ := c.Get("username")
		if userID == nil {
			return // skip unauthenticated requests
		}

		log := &model.AuditLog{
			UserID:     userID.(int64),
			Username:   username.(string),
			Action:     c.Request.Method,
			Resource:   c.FullPath(),
			ResourceID: c.Param("id"),
			IP:         c.ClientIP(),
			UserAgent:  c.Request.UserAgent(),
			Duration:   int(duration),
			StatusCode: c.Writer.Status(),
		}

		// Truncate body for detail
		if len(bodyBytes) > 500 {
			bodyBytes = bodyBytes[:500]
		}
		log.Detail = string(bodyBytes)

		if err := auditRepo.Create(c.Request.Context(), log); err != nil {
			logger.Error("Failed to write audit log: %v", err)
		}
	}
}
