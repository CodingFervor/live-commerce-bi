package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/CodingFervor/live-commerce-bi/internal/database"
	"github.com/CodingFervor/live-commerce-bi/pkg/response"

	"github.com/gin-gonic/gin"
)

// ═══ Row-Level Data Permission Middleware ═══
// Injects data scope filters based on user's role-based data permissions
// Reference: Alibaba fine-grained data access control

// DataScope holds the data scope for the current user
type DataScope struct {
	UserID         int64    `json:"user_id"`
	OrganizationID int64    `json:"organization_id"`
	DepartmentIDs  []int64  `json:"department_ids"`
	PlatformFilter []string `json:"platform_filter"`
	StreamerFilter []int64  `json:"streamer_filter"`
	IsFullAccess   bool     `json:"is_full_access"`
}

const dataScopeKey = "data_scope"

// DataPermission injects data scope into context for row-level filtering
func DataPermission() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.Next()
			return
		}

		role, _ := c.Get("role")
		if role == "admin" {
			// Admin has full access
			c.Set(dataScopeKey, &DataScope{
				UserID:       userID.(int64),
				IsFullAccess: true,
			})
			c.Next()
			return
		}

		// Load data permissions for the user
		scope := loadDataScope(c.Request.Context(), userID.(int64))
		c.Set(dataScopeKey, scope)
		c.Next()
	}
}

// GetDataScope extracts the data scope from gin context
func GetDataScope(c *gin.Context) *DataScope {
	scope, exists := c.Get(dataScopeKey)
	if !exists {
		return &DataScope{IsFullAccess: false}
	}
	return scope.(*DataScope)
}

// ApplyDataFilter generates SQL WHERE clause for data permission filtering
func (ds *DataScope) ApplyDataFilter(tableName string) string {
	if ds.IsFullAccess {
		return ""
	}

	var conditions []string

	if len(ds.PlatformFilter) > 0 {
		platforms := make([]string, len(ds.PlatformFilter))
		for i, p := range ds.PlatformFilter {
			// Escape single quotes to prevent SQL injection
			platforms[i] = fmt.Sprintf("'%s'", strings.ReplaceAll(p, "'", "''"))
		}
		conditions = append(conditions,
			fmt.Sprintf("%s.platform IN (%s)", tableName, strings.Join(platforms, ",")))
	}

	if len(ds.StreamerFilter) > 0 {
		streamers := make([]string, len(ds.StreamerFilter))
		for i, s := range ds.StreamerFilter {
			streamers[i] = fmt.Sprintf("%d", s)
		}
		conditions = append(conditions,
			fmt.Sprintf("%s.streamer_id IN (%s)", tableName, strings.Join(streamers, ",")))
	}

	if len(conditions) == 0 {
		return ""
	}

	return " AND " + strings.Join(conditions, " AND ")
}

// RequireDataScope is a middleware that blocks access if user has no data scope
func RequireDataScope() gin.HandlerFunc {
	return func(c *gin.Context) {
		scope := GetDataScope(c)
		if scope == nil || (!scope.IsFullAccess && len(scope.PlatformFilter) == 0 && len(scope.StreamerFilter) == 0) {
			response.Forbidden(c, "no data access permissions")
			c.Abort()
			return
		}
		c.Next()
	}
}

func loadDataScope(ctx context.Context, userID int64) *DataScope {
	scope := &DataScope{
		UserID:       userID,
		IsFullAccess: false,
	}

	// Query user's data permissions from database
	rows, err := database.Get().Query(ctx, `
		SELECT dp.dimension, dp.values
		FROM data_permissions dp
		JOIN user_roles ur ON dp.role_id = ur.role_id
		WHERE ur.user_id = $1`, userID)
	if err != nil {
		return scope
	}
	defer rows.Close()

	for rows.Next() {
		var dimension, values string
		if rows.Scan(&dimension, values) != nil {
			continue
		}

		switch dimension {
		case "platform":
			// values is JSON array like ["douyin","kuaishou"]
			parseJSONArray(values, &scope.PlatformFilter)
		case "streamer":
			// values is JSON array of streamer IDs
			var ids []int64
			parseJSONArray(values, &ids)
			scope.StreamerFilter = ids
		case "department":
			parseJSONArray(values, &scope.DepartmentIDs)
		case "full_access":
			scope.IsFullAccess = values == "true" || values == "1"
		}
	}

	// Get user's organization
	database.Get().QueryRow(ctx,
		"SELECT organization_id FROM users WHERE id=$1", userID).Scan(&scope.OrganizationID)

	return scope
}
