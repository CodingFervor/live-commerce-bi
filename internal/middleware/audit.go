package middleware

import (
	"context"
	"time"

	"github.com/CodingFervor/live-commerce-bi/internal/model"
	"github.com/CodingFervor/live-commerce-bi/internal/repository"
	"github.com/CodingFervor/live-commerce-bi/pkg/logger"

	"github.com/gin-gonic/gin"
)

// AuditLogger records all API operations to audit_logs table
func AuditLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start).Milliseconds()

		userID, _ := c.Get("user_id")
		username, _ := c.Get("username")
		if userID == nil {
			return
		}

		log := &model.AuditLog{
			UserID:     userID.(int64),
			Username:   username.(string),
			Action:     c.Request.Method,
			Resource:   c.FullPath(),
			ResourceID: c.Param("id"),
			IP:         c.ClientIP(),
			UserAgent:  c.Request.UserAgent(),
			RequestID:  c.GetString("request_id"),
			Duration:   int(duration),
			StatusCode: c.Writer.Status(),
		}

		auditRepo := repository.NewAuditRepo()
		if err := auditRepo.Create(c.Request.Context(), log); err != nil {
			logger.Error("Failed to write audit log: %v", err)
		}
	}
}

// RequestID injects a unique request ID into context
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

func generateRequestID() string {
	return time.Now().Format("20060102150405") + "-" + randomHex(8)
}

func randomHex(n int) string {
	const hex = "0123456789abcdef"
	b := make([]byte, n)
	for i := range b {
		b[i] = hex[(time.Now().UnixNano()+int64(i))%16]
	}
	return string(b)
}

// suppress unused import
var _ = context.Background
