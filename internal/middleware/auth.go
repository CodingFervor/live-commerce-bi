package middleware

import (
	"net/http"
	"strings"

	"github.com/CodingFervor/live-commerce-bi/pkg/jwt"
	"github.com/CodingFervor/live-commerce-bi/pkg/response"

	"github.com/gin-gonic/gin"
)

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "missing authorization header")
			c.Abort()
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "invalid authorization format")
			c.Abort()
			return
		}
		claims, err := jwt.ParseToken(parts[1])
		if err != nil {
			response.Unauthorized(c, "invalid or expired token")
			c.Abort()
			return
		}
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		if role != "admin" {
			response.Forbidden(c, "admin access required")
			c.Abort()
			return
		}
		c.Next()
	}
}

func RoleRequired(roles ...string) gin.HandlerFunc {
	roleSet := make(map[string]bool)
	for _, r := range roles {
		roleSet[r] = true
	}
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		if !roleSet[role.(string)] {
			response.Forbidden(c, "insufficient permissions")
			c.Abort()
			return
		}
		c.Next()
	}
}

func GetUserID(c *gin.Context) int64 {
	id, _ := c.Get("user_id")
	if id == nil {
		return 0
	}
	return id.(int64)
}

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS,PATCH")
		c.Header("Access-Control-Allow-Headers", "Authorization,Content-Type,X-Requested-With")
		c.Header("Access-Control-Max-Age", "86400")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
