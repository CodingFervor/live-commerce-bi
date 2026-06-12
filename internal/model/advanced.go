package model

import "time"

// ═══════════════════════════════════════════════════════════
// ADVANCED MODELS - Enterprise BI Enhancement
// Reference: ByteDance/Alibaba Data Platform Architecture
// ═══════════════════════════════════════════════════════════

// ─── Organization & Department (Multi-tenant) ───
type Organization struct {
	ID          int64      `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	Code        string     `json:"code" db:"code"`
	Industry    string     `json:"industry" db:"industry"`
	Plan        string     `json:"plan" db:"plan"`
	MaxUsers    int        `json:"max_users" db:"max_users"`
	MaxRooms    int        `json:"max_rooms" db:"max_rooms"`
	Status      string     `json:"status" db:"status"`
	ExpiredAt   *time.Time `json:"expired_at" db:"expired_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

type Department struct {
	ID             int64        `json:"id" db:"id"`
	OrganizationID int64        `json:"organization_id" db:"organization_id"`
	ParentID       *int64       `json:"parent_id" db:"parent_id"`
	Name           string       `json:"name" db:"name"`
	Code           string       `json:"code" db:"code"`
	LeaderID       *int64       `json:"leader_id" db:"leader_id"`
	SortOrder      int          `json:"sort_order" db:"sort_order"`
	Path           string       `json:"path" db:"path"`
	Level          int          `json:"level" db:"level"`
	Status         string       `json:"status" db:"status"`
	CreatedAt      time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at" db:"updated_at"`
	Children       []Department `json:"children,omitempty" db:"-"`
}

// ─── Fine-grained RBAC ───
type Permission struct {
	ID       int64  `json:"id" db:"id"`
	Code     string `json:"code" db:"code"`
	Name     string `json:"name" db:"name"`
	Module   string `json:"module" db:"module"`
	Resource string `json:"resource" db:"resource"`
	Action   string `json:"action" db:"action"`
}

type Role struct {
	ID             int64        `json:"id" db:"id"`
	OrganizationID int64        `json:"organization_id" db:"organization_id"`
	Name           string       `json:"name" db:"name"`
	Code           string       `json:"code" db:"code"`
	IsSystem       bool         `json:"is_system" db:"is_system"`
	Description    string       `json:"description" db:"description"`
	CreatedAt      time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at" db:"updated_at"`
	Permissions    []Permission `json:"permissions,omitempty" db:"-"`
}

type UserRole struct {
	ID     int64 `json:"id" db:"id"`
	UserID int64 `json:"user_id" db:"user_id"`
	RoleID int64 `json:"role_id" db:"role_id"`
}

type DataPermission struct {
	ID        int64  `json:"id" db:"id"`
	RoleID    int64  `json:"role_id" db:"role_id"`
	Dimension string `json:"dimension" db:"dimension"`
	Values    string `json:"values" db:"values"`
}

// ─── Audit Log ───
type AuditLog struct {
	ID         int64     `json:"id" db:"id"`
	UserID     int64     `json:"user_id" db:"user_id"`
	Username   string    `json:"username" db:"username"`
	Action     string    `json:"action" db:"action"`
	Resource   string    `json:"resource" db:"resource"`
	ResourceID string    `json:"resource_id" db:"resource_id"`
	Detail     string    `json:"detail" db:"detail"`
	IP         string    `json:"ip" db:"ip"`
	UserAgent  string    `json:"user_agent" db:"user_agent"`
	RequestID  string    `json:"request_id" db:"request_id"`
	Duration   int       `json:"duration" db:"duration"`
	StatusCode int       `json:"status_code" db:"status_code"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// ─── Event Tracking ───
type TrackEvent struct {
	ID         int64     `json:"id" db:"id"`
	EventName  string    `json:"event_name" db:"event_name"`
	Platform   string    `json:"platform" db:"platform"`
	UserID     string    `json:"user_id" db:"user_id"`
	SessionID  string    `json:"session_id" db:"session_id"`
	LiveRoomID int64     `json:"live_room_id" db:"live_room_id"`
	ProductID  int64     `json:"product_id" db:"product_id"`
	Properties string    `json:"properties" db:"properties"`
	Timestamp  time.Time `json:"timestamp" db:"timestamp"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

type UserBehaviorPath struct {
	SessionID string   `json:"session_id"`
	UserID    string   `json:"user_id"`
	Events    []string `json:"events"`
	Duration  int      `json:"duration"`
	Converted bool     `json:"converted"`
}

type RFMSegment struct {
	UserID    string  `json:"user_id"`
	Recency   int     `json:"recency"`
	Frequency int     `json:"frequency"`
	Monetary  float64 `json:"monetary"`
	Score     int     `json:"score"`
	Segment   string  `json:"segment"`
}

// ─── Export Task ───
type ExportTask struct {
	ID          int64      `json:"id" db:"id"`
	UserID      int64      `json:"user_id" db:"user_id"`
	Type        string     `json:"type" db:"type"`
	Format      string     `json:"format" db:"format"`
	Params      string     `json:"params" db:"params"`
	Status      string     `json:"status" db:"status"`
	Progress    int        `json:"progress" db:"progress"`
	TotalRows   int        `json:"total_rows" db:"total_rows"`
	FilePath    string     `json:"file_path" db:"file_path"`
	FileSize    int64      `json:"file_size" db:"file_size"`
	ErrorMsg    string     `json:"error_msg" db:"error_msg"`
	StartedAt   *time.Time `json:"started_at" db:"started_at"`
	CompletedAt *time.Time `json:"completed_at" db:"completed_at"`
	ExpiresAt   *time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}

// ─── Data Quality ───
type DataQualityRule struct {
	ID          int64      `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	TableName   string     `json:"table_name" db:"table_name"`
	ColumnName  string     `json:"column_name" db:"column_name"`
	RuleType    string     `json:"rule_type" db:"rule_type"`
	Expression  string     `json:"expression" db:"expression"`
	Severity    string     `json:"severity" db:"severity"`
	IsEnabled   bool       `json:"is_enabled" db:"is_enabled"`
	LastCheckAt *time.Time `json:"last_check_at" db:"last_check_at"`
	CreatedBy   int64      `json:"created_by" db:"created_by"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

type DataQualityResult struct {
	ID        int64     `json:"id" db:"id"`
	RuleID    int64     `json:"rule_id" db:"rule_id"`
	Status    string    `json:"status" db:"status"`
	TotalRows int       `json:"total_rows" db:"total_rows"`
	PassRows  int       `json:"pass_rows" db:"pass_rows"`
	FailRows  int       `json:"fail_rows" db:"fail_rows"`
	PassRate  float64   `json:"pass_rate" db:"pass_rate"`
	Detail    string    `json:"detail" db:"detail"`
	CheckedAt time.Time `json:"checked_at" db:"checked_at"`
}

// ─── Scheduled Task ───
type ScheduledTask struct {
	ID        int64      `json:"id" db:"id"`
	Name      string     `json:"name" db:"name"`
	Type      string     `json:"type" db:"type"`
	CronExpr  string     `json:"cron_expr" db:"cron_expr"`
	Config    string     `json:"config" db:"config"`
	Status    string     `json:"status" db:"status"`
	LastRunAt *time.Time `json:"last_run_at" db:"last_run_at"`
	NextRunAt *time.Time `json:"next_run_at" db:"next_run_at"`
	RunCount  int        `json:"run_count" db:"run_count"`
	LastError string     `json:"last_error" db:"last_error"`
	CreatedBy int64      `json:"created_by" db:"created_by"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
}

// ─── Notification ───
type NotificationTemplate struct {
	ID        int64     `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Channel   string    `json:"channel" db:"channel"`
	Subject   string    `json:"subject" db:"subject"`
	Content   string    `json:"content" db:"content"`
	IsDefault bool      `json:"is_default" db:"is_default"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type NotificationRecord struct {
	ID         int64      `json:"id" db:"id"`
	UserID     int64      `json:"user_id" db:"user_id"`
	TemplateID int64      `json:"template_id" db:"template_id"`
	Channel    string     `json:"channel" db:"channel"`
	Recipient  string     `json:"recipient" db:"recipient"`
	Subject    string     `json:"subject" db:"subject"`
	Content    string     `json:"content" db:"content"`
	Status     string     `json:"status" db:"status"`
	ProviderID string     `json:"provider_id" db:"provider_id"`
	RetryCount int        `json:"retry_count" db:"retry_count"`
	SentAt     *time.Time `json:"sent_at" db:"sent_at"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
}

// ─── Metrics Aggregation ───
type MetricsHourly struct {
	ID             int64     `json:"id" db:"id"`
	Platform       string    `json:"platform" db:"platform"`
	Hour           time.Time `json:"hour" db:"hour"`
	StreamerID     int64     `json:"streamer_id" db:"streamer_id"`
	LiveRoomID     int64     `json:"live_room_id" db:"live_room_id"`
	GMV            float64   `json:"gmv" db:"gmv"`
	OrderCount     int       `json:"order_count" db:"order_count"`
	ViewerCount    int64     `json:"viewer_count" db:"viewer_count"`
	PeakViewers    int       `json:"peak_viewers" db:"peak_viewers"`
	LikeCount      int64     `json:"like_count" db:"like_count"`
	CommentCount   int64     `json:"comment_count" db:"comment_count"`
	ShareCount     int64     `json:"share_count" db:"share_count"`
	ConversionRate float64   `json:"conversion_rate" db:"conversion_rate"`
	AvgWatchTime   int       `json:"avg_watch_time" db:"avg_watch_time"`
	NewFollowers   int       `json:"new_followers" db:"new_followers"`
	GiftCount      int       `json:"gift_count" db:"gift_count"`
	GiftValue      float64   `json:"gift_value" db:"gift_value"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

type MetricsDaily struct {
	ID             int64     `json:"id" db:"id"`
	Platform       string    `json:"platform" db:"platform"`
	Date           time.Time `json:"date" db:"date"`
	StreamerID     int64     `json:"streamer_id" db:"streamer_id"`
	LiveRoomCount  int       `json:"live_room_count" db:"live_room_count"`
	TotalDuration  int       `json:"total_duration" db:"total_duration"`
	GMV            float64   `json:"gmv" db:"gmv"`
	OrderCount     int       `json:"order_count" db:"order_count"`
	ViewerCount    int64     `json:"viewer_count" db:"viewer_count"`
	PeakViewers    int       `json:"peak_viewers" db:"peak_viewers"`
	LikeCount      int64     `json:"like_count" db:"like_count"`
	CommentCount   int64     `json:"comment_count" db:"comment_count"`
	ShareCount     int64     `json:"share_count" db:"share_count"`
	ConversionRate float64   `json:"conversion_rate" db:"conversion_rate"`
	AvgWatchTime   int       `json:"avg_watch_time" db:"avg_watch_time"`
	NewFollowers   int       `json:"new_followers" db:"new_followers"`
	GiftCount      int       `json:"gift_count" db:"gift_count"`
	GiftValue      float64   `json:"gift_value" db:"gift_value"`
	RefundAmount   float64   `json:"refund_amount" db:"refund_amount"`
	Commission     float64   `json:"commission" db:"commission"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// ─── OLAP / Analysis Models ───
type OLAPQuery struct {
	Dimensions []string          `json:"dimensions"`
	Metrics    []string          `json:"metrics"`
	Filters    map[string]string `json:"filters"`
	OrderBy    string            `json:"order_by"`
	OrderDir   string            `json:"order_dir"`
	Limit      int               `json:"limit"`
}

type OLAPResult struct {
	Dimensions map[string]interface{} `json:"dimensions"`
	Metrics    map[string]interface{} `json:"metrics"`
}

type ComparisonResult struct {
	Current   map[string]interface{} `json:"current"`
	Previous  map[string]interface{} `json:"previous"`
	Change    map[string]interface{} `json:"change"`
	ChangePct map[string]interface{} `json:"change_pct"`
}

type SalesForecast struct {
	Date     string  `json:"date"`
	Actual   float64 `json:"actual,omitempty"`
	Forecast float64 `json:"forecast"`
	Lower    float64 `json:"lower"`
	Upper    float64 `json:"upper"`
}

type AnomalyPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
	Expected  float64   `json:"expected"`
	Deviation float64   `json:"deviation"`
	IsAnomaly bool      `json:"is_anomaly"`
	Severity  string    `json:"severity"`
}

type DashboardShare struct {
	ID           int64      `json:"id" db:"id"`
	DashboardID  int64      `json:"dashboard_id" db:"dashboard_id"`
	ShareToken   string     `json:"share_token" db:"share_token"`
	Password     string     `json:"password" db:"password"`
	AllowedRoles string     `json:"allowed_roles" db:"allowed_roles"`
	ExpiresAt    *time.Time `json:"expires_at" db:"expires_at"`
	ViewCount    int        `json:"view_count" db:"view_count"`
	CreatedBy    int64      `json:"created_by" db:"created_by"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
}

type LiveRoomTimeline struct {
	ID         int64     `json:"id" db:"id"`
	LiveRoomID int64     `json:"live_room_id" db:"live_room_id"`
	Timestamp  time.Time `json:"timestamp" db:"timestamp"`
	EventType  string    `json:"event_type" db:"event_type"`
	Title      string    `json:"title" db:"title"`
	Data       string    `json:"data" db:"data"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}
