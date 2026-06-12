package model

import "time"

// ═══ AI Configuration Models ═══
// Multi-model AI integration for intelligent BI analysis

type AIProvider string

const (
	AIProviderOpenAI    AIProvider = "openai"
	AIProviderQwen      AIProvider = "qwen"       // 通义千问
	AIProviderZhipu     AIProvider = "zhipu"      // 智谱 GLM
	AIProviderBaidu     AIProvider = "baidu"      // 百度文心
	AIProviderDeepSeek  AIProvider = "deepseek"
	AIProviderMoonshot  AIProvider = "moonshot"   // 月之暗面 Kimi
	AIProviderSpark     AIProvider = "spark"      // 讯飞星火
	AIProviderOllama    AIProvider = "ollama"     // 本地部署
)

// AIConfig stores an AI model provider configuration
type AIConfig struct {
	ID          int64     `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Provider    string    `json:"provider" db:"provider"`
	APIKey      string    `json:"api_key,omitempty" db:"api_key"`
	APIEndpoint string    `json:"api_endpoint" db:"api_endpoint"`
	ModelName   string    `json:"model_name" db:"model_name"`
	MaxTokens   int       `json:"max_tokens" db:"max_tokens"`
	Temperature float64   `json:"temperature" db:"temperature"`
	TopP        float64   `json:"top_p" db:"top_p"`
	IsDefault   bool      `json:"is_default" db:"is_default"`
	IsEnabled   bool      `json:"is_enabled" db:"is_enabled"`
	ProxyURL    string    `json:"proxy_url,omitempty" db:"proxy_url"`
	ExtraConfig string    `json:"extra_config,omitempty" db:"extra_config"`
	CreatedBy   int64     `json:"created_by" db:"created_by"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// MaskedAPIKey returns the API key with only last 4 characters visible
func (c *AIConfig) MaskedAPIKey() string {
	if c.APIKey == "" {
		return ""
	}
	if len(c.APIKey) <= 8 {
		return "****"
	}
	return "sk-****" + c.APIKey[len(c.APIKey)-4:]
}

// ToPublic returns a safe copy with masked API key for API responses
func (c *AIConfig) ToPublic() map[string]interface{} {
	return map[string]interface{}{
		"id":           c.ID,
		"name":         c.Name,
		"provider":     c.Provider,
		"api_key":      c.MaskedAPIKey(),
		"api_endpoint": c.APIEndpoint,
		"model_name":   c.ModelName,
		"max_tokens":   c.MaxTokens,
		"temperature":  c.Temperature,
		"top_p":        c.TopP,
		"is_default":   c.IsDefault,
		"is_enabled":   c.IsEnabled,
		"proxy_url":    c.ProxyURL,
		"extra_config": c.ExtraConfig,
		"created_by":   c.CreatedBy,
		"created_at":   c.CreatedAt,
		"updated_at":   c.UpdatedAt,
	}
}

type AIConfigCreate struct {
	Name        string  `json:"name" binding:"required"`
	Provider    string  `json:"provider" binding:"required,oneof=openai qwen zhipu baidu deepseek moonshot spark ollama"`
	APIKey      string  `json:"api_key"`
	APIEndpoint string  `json:"api_endpoint"`
	ModelName   string  `json:"model_name" binding:"required"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
	TopP        float64 `json:"top_p"`
	IsDefault   bool    `json:"is_default"`
	ProxyURL    string  `json:"proxy_url"`
	ExtraConfig string  `json:"extra_config"`
}

type AIConfigUpdate struct {
	Name        *string  `json:"name"`
	APIKey      *string  `json:"api_key"`
	APIEndpoint *string  `json:"api_endpoint"`
	ModelName   *string  `json:"model_name"`
	MaxTokens   *int     `json:"max_tokens"`
	Temperature *float64 `json:"temperature"`
	TopP        *float64 `json:"top_p"`
	IsDefault   *bool    `json:"is_default"`
	IsEnabled   *bool    `json:"is_enabled"`
	ProxyURL    *string  `json:"proxy_url"`
	ExtraConfig *string  `json:"extra_config"`
}

// AIConversation represents an AI chat session
type AIConversation struct {
	ID         int64     `json:"id" db:"id"`
	UserID     int64     `json:"user_id" db:"user_id"`
	Title      string    `json:"title" db:"title"`
	ConfigID   int64     `json:"config_id" db:"config_id"`
	Messages   string    `json:"messages" db:"messages"` // JSON array of messages
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

type AIMessage struct {
	Role    string `json:"role"`    // system, user, assistant
	Content string `json:"content"`
}

type AIChatRequest struct {
	ConfigID    int64       `json:"config_id"`
	ConversationID int64    `json:"conversation_id"`
	Message     string      `json:"message" binding:"required"`
	Context     *AIContext  `json:"context,omitempty"`
}

type AIContext struct {
	DataType   string `json:"data_type"`   // gmv, orders, streamers, etc.
	DateRange  string `json:"date_range"`
	RoomID     int64  `json:"room_id"`
	Platform   string `json:"platform"`
}

type AIChatResponse struct {
	ConversationID int64      `json:"conversation_id"`
	Reply          string     `json:"reply"`
	SQL            string     `json:"sql,omitempty"`
	Data           interface{} `json:"data,omitempty"`
	ChartSuggestion string    `json:"chart_suggestion,omitempty"`
	TokensUsed     int        `json:"tokens_used"`
}

type AIInsightRequest struct {
	ConfigID  int64  `json:"config_id"`
	Topic     string `json:"topic" binding:"required,oneof=gmv conversion streamer product viewer retention anomaly forecast"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Platform  string `json:"platform"`
}

type AIInsightResponse struct {
	Topic       string      `json:"topic"`
	Summary     string      `json:"summary"`
	Insights    []string    `json:"insights"`
	Recommendations []string `json:"recommendations"`
	Data        interface{} `json:"data,omitempty"`
	TokensUsed  int         `json:"tokens_used"`
}

type AITestRequest struct {
	ConfigID int64  `json:"config_id"`
	Prompt   string `json:"prompt"`
}

type AITestResponse struct {
	Success    bool   `json:"success"`
	Reply      string `json:"reply"`
	Error      string `json:"error,omitempty"`
	TokensUsed int    `json:"tokens_used"`
	LatencyMs  int64  `json:"latency_ms"`
}

// ═══ System Settings Models ═══

type SystemSetting struct {
	ID        int64     `json:"id" db:"id"`
	Category  string    `json:"category" db:"category"`
	Key       string    `json:"key" db:"key"`
	Value     string    `json:"value" db:"value"`
	ValueType string    `json:"value_type" db:"value_type"` // string, int, float, bool, json
	Remark    string    `json:"remark" db:"remark"`
	IsPublic  bool      `json:"is_public" db:"is_public"`   // visible to non-admin
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type SystemSettingUpdate struct {
	Value string `json:"value" binding:"required"`
	Remark string `json:"remark"`
}

type SMTPConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	FromName string `json:"from_name"`
	FromAddr string `json:"from_addr"`
	UseTLS   bool   `json:"use_tls"`
}

type StorageConfig struct {
	Provider string `json:"provider"` // local, oss, s3, minio
	Endpoint string `json:"endpoint"`
	Bucket   string `json:"bucket"`
	AK       string `json:"access_key"`
	SK       string `json:"secret_key"`
	Region   string `json:"region"`
	PathPrefix string `json:"path_prefix"`
}

type SecurityConfig struct {
	PasswordMinLength int    `json:"password_min_length"`
	PasswordRequireUpper bool `json:"password_require_upper"`
	PasswordRequireNumber bool `json:"password_require_number"`
	PasswordRequireSpecial bool `json:"password_require_special"`
	LoginMaxAttempts  int    `json:"login_max_attempts"`
	LoginLockDuration int    `json:"login_lock_duration"` // minutes
	SessionTimeout    int    `json:"session_timeout"`     // hours
	IPWhitelist       string `json:"ip_whitelist"`        // comma separated
	Enable2FA         bool   `json:"enable_2fa"`
}

type SystemInfo struct {
	Version       string `json:"version"`
	GoVersion     string `json:"go_version"`
	StartTime     string `json:"start_time"`
	Uptime        string `json:"uptime"`
	TotalUsers    int64  `json:"total_users"`
	TotalRooms    int64  `json:"total_rooms"`
	TotalStreamers int64 `json:"total_streamers"`
	TotalOrders   int64  `json:"total_orders"`
	DBSize        string `json:"db_size"`
	RedisMemory   string `json:"redis_memory"`
}
