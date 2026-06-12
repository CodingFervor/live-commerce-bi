package handler

import (
	"encoding/json"
	"strconv"

	"github.com/CodingFervor/live-commerce-bi/internal/middleware"
	"github.com/CodingFervor/live-commerce-bi/internal/model"
	"github.com/CodingFervor/live-commerce-bi/internal/repository"
	"github.com/CodingFervor/live-commerce-bi/internal/service"
	"github.com/CodingFervor/live-commerce-bi/pkg/response"

	"github.com/gin-gonic/gin"
)

// ═══ AI Config Handler ═══

type AIConfigHandler struct {
	repo    *repository.AIConfigRepo
	aiSvc   *service.AIService
}

func NewAIConfigHandler() *AIConfigHandler {
	return &AIConfigHandler{
		repo:  repository.NewAIConfigRepo(),
		aiSvc: service.NewAIService(),
	}
}

// ListAIConfigs returns all AI configurations (API keys masked)
// GET /api/v1/ai/configs
func (h *AIConfigHandler) ListAIConfigs(c *gin.Context) {
	list, err := h.repo.List(c.Request.Context())
	if err != nil {
		response.InternalError(c, "failed to list AI configs")
		return
	}
	// Mask API keys in response
	masked := make([]map[string]interface{}, len(list))
	for i, cfg := range list {
		masked[i] = cfg.ToPublic()
	}
	response.OK(c, masked)
}

// GetAIConfig returns a single AI config (API key masked)
// GET /api/v1/ai/configs/:id
func (h *AIConfigHandler) GetAIConfig(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	cfg, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "AI config not found")
		return
	}
	response.OK(c, cfg.ToPublic())
}

// CreateAIConfig creates a new AI provider configuration
// POST /api/v1/ai/configs
func (h *AIConfigHandler) CreateAIConfig(c *gin.Context) {
	var req model.AIConfigCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	cfg := &model.AIConfig{
		Name:        req.Name,
		Provider:    req.Provider,
		APIKey:      req.APIKey,
		APIEndpoint: req.APIEndpoint,
		ModelName:   req.ModelName,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		IsDefault:   req.IsDefault,
		IsEnabled:   true,
		ProxyURL:    req.ProxyURL,
		ExtraConfig: req.ExtraConfig,
		CreatedBy:   middleware.GetUserID(c),
	}
	if cfg.MaxTokens == 0 {
		cfg.MaxTokens = 4096
	}
	if cfg.Temperature == 0 {
		cfg.Temperature = 0.7
	}

	if err := h.repo.Create(c.Request.Context(), cfg); err != nil {
		response.InternalError(c, "failed to create AI config")
		return
	}
	response.Created(c, cfg.ToPublic())
}

// UpdateAIConfig updates an AI configuration
// PUT /api/v1/ai/configs/:id
func (h *AIConfigHandler) UpdateAIConfig(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req model.AIConfigUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	cfg, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "AI config not found")
		return
	}

	if req.Name != nil { cfg.Name = *req.Name }
	if req.APIKey != nil { cfg.APIKey = *req.APIKey }
	if req.APIEndpoint != nil { cfg.APIEndpoint = *req.APIEndpoint }
	if req.ModelName != nil { cfg.ModelName = *req.ModelName }
	if req.MaxTokens != nil { cfg.MaxTokens = *req.MaxTokens }
	if req.Temperature != nil { cfg.Temperature = *req.Temperature }
	if req.TopP != nil { cfg.TopP = *req.TopP }
	if req.IsDefault != nil { cfg.IsDefault = *req.IsDefault }
	if req.IsEnabled != nil { cfg.IsEnabled = *req.IsEnabled }
	if req.ProxyURL != nil { cfg.ProxyURL = *req.ProxyURL }
	if req.ExtraConfig != nil { cfg.ExtraConfig = *req.ExtraConfig }

	if err := h.repo.Update(c.Request.Context(), cfg); err != nil {
		response.InternalError(c, "failed to update AI config")
		return
	}
	response.OK(c, cfg.ToPublic())
}

// DeleteAIConfig deletes an AI configuration
// DELETE /api/v1/ai/configs/:id
func (h *AIConfigHandler) DeleteAIConfig(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		response.InternalError(c, "failed to delete AI config")
		return
	}
	response.OKMsg(c, "deleted")
}

// TestAIConfig tests the connection to an AI provider
// POST /api/v1/ai/configs/:id/test
func (h *AIConfigHandler) TestAIConfig(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	cfg, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "AI config not found")
		return
	}
	result, err := h.aiSvc.TestConnection(c.Request.Context(), cfg)
	if err != nil {
		response.OK(c, result)
		return
	}
	response.OK(c, result)
}

// SetDefaultAIConfig sets an AI config as the default
// POST /api/v1/ai/configs/:id/default
func (h *AIConfigHandler) SetDefaultAIConfig(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.repo.SetDefault(c.Request.Context(), id); err != nil {
		response.InternalError(c, "failed to set default AI config")
		return
	}
	response.OKMsg(c, "default AI config updated")
}

// ═══ AI Chat Handler ═══

type AIChatHandler struct {
	convRepo *repository.AIConversationRepo
	aiSvc    *service.AIService
}

func NewAIChatHandler() *AIChatHandler {
	return &AIChatHandler{
		convRepo: repository.NewAIConversationRepo(),
		aiSvc:    service.NewAIService(),
	}
}

// Chat handles AI chat requests
// POST /api/v1/ai/chat
func (h *AIChatHandler) Chat(c *gin.Context) {
	var req model.AIChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Load AI config
	var cfg *model.AIConfig
	var err error
	if req.ConfigID > 0 {
		cfg, err = h.aiSvc.LoadConfig(c.Request.Context(), req.ConfigID)
	} else {
		cfg, err = h.aiSvc.LoadDefaultConfig(c.Request.Context())
	}
	if err != nil {
		response.BadRequest(c, "AI config not available")
		return
	}

	// Load or create conversation
	var messages []model.AIMessage
	if req.ConversationID > 0 {
		conv, err := h.convRepo.GetByID(c.Request.Context(), req.ConversationID)
		if err == nil {
			// Ownership check: only the conversation owner can continue it
			if conv.UserID != middleware.GetUserID(c) {
				response.Forbidden(c, "access denied to this conversation")
				return
			}
			json.Unmarshal([]byte(conv.Messages), &messages)
		}
	}

	messages = append(messages, model.AIMessage{Role: "user", Content: req.Message})

	// Call AI
	aiResp, err := h.aiSvc.SmartQuery(c.Request.Context(), cfg, req.Message, req.Context)
	if err != nil {
		response.InternalError(c, "AI request failed")
		return
	}

	// Save conversation
	messages = append(messages, model.AIMessage{Role: "assistant", Content: aiResp.Reply})
	msgJSON, _ := json.Marshal(messages)

	if req.ConversationID > 0 {
		conv, _ := h.convRepo.GetByID(c.Request.Context(), req.ConversationID)
		if conv != nil {
			conv.Messages = string(msgJSON)
			h.convRepo.Update(c.Request.Context(), conv)
			aiResp.ConversationID = conv.ID
		}
	} else {
		conv := &model.AIConversation{
			UserID:   middleware.GetUserID(c),
			Title:    req.Message[:min(len(req.Message), 50)],
			ConfigID: cfg.ID,
			Messages: string(msgJSON),
		}
		if err := h.convRepo.Create(c.Request.Context(), conv); err == nil {
			aiResp.ConversationID = conv.ID
		}
	}

	response.OK(c, aiResp)
}

// ListConversations returns user's AI chat conversations
// GET /api/v1/ai/conversations
func (h *AIChatHandler) ListConversations(c *gin.Context) {
	userID := middleware.GetUserID(c)
	list, err := h.convRepo.ListByUser(c.Request.Context(), userID, 20)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, list)
}

// GetConversation returns a conversation with messages (owner only)
// GET /api/v1/ai/conversations/:id
func (h *AIChatHandler) GetConversation(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	conv, err := h.convRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "conversation not found")
		return
	}
	// Ownership check
	if conv.UserID != middleware.GetUserID(c) {
		response.Forbidden(c, "access denied")
		return
	}
	response.OK(c, conv)
}

// DeleteConversation deletes a chat conversation (owner only)
// DELETE /api/v1/ai/conversations/:id
func (h *AIChatHandler) DeleteConversation(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	conv, err := h.convRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "conversation not found")
		return
	}
	// Ownership check
	if conv.UserID != middleware.GetUserID(c) {
		response.Forbidden(c, "access denied")
		return
	}
	if err := h.convRepo.Delete(c.Request.Context(), id); err != nil {
		response.InternalError(c, "failed to delete conversation")
		return
	}
	response.OKMsg(c, "deleted")
}

// GenerateInsights generates AI insights for a topic
// POST /api/v1/ai/insights
func (h *AIChatHandler) GenerateInsights(c *gin.Context) {
	var req model.AIInsightRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	cfg, err := h.aiSvc.LoadDefaultConfig(c.Request.Context())
	if err != nil {
		response.BadRequest(c, "AI config not available")
		return
	}

	// Get analytics data for context
	analyticsData := h.getTopicData(c, req.Topic, req.StartDate, req.EndDate, req.Platform)

	result, err := h.aiSvc.GenerateInsights(c.Request.Context(), cfg, &req, analyticsData)
	if err != nil {
		response.InternalError(c, "AI insight generation failed")
		return
	}
	response.OK(c, result)
}

func (h *AIChatHandler) getTopicData(c *gin.Context, topic, startDate, endDate, platform string) map[string]interface{} {
	return map[string]interface{}{
		"topic": topic, "start_date": startDate, "end_date": endDate,
		"platform": platform, "note": "data loaded from analytics service",
	}
}

// ═══ System Settings Handler ═══

type SystemSettingHandler struct {
	repo *repository.SystemSettingRepo
}

func NewSystemSettingHandler() *SystemSettingHandler {
	return &SystemSettingHandler{repo: repository.NewSystemSettingRepo()}
}

// ListSettings returns all system settings grouped by category
// GET /api/v1/settings?category=smtp
func (h *SystemSettingHandler) ListSettings(c *gin.Context) {
	category := c.Query("category")
	if category != "" {
		list, err := h.repo.ListByCategory(c.Request.Context(), category)
		if err != nil {
			response.InternalError(c, "failed to load settings")
			return
		}
		response.OK(c, list)
		return
	}

	all, err := h.repo.ListAll(c.Request.Context())
	if err != nil {
		response.InternalError(c, "failed to load settings")
		return
	}
	// Group by category, mask sensitive values
	grouped := make(map[string][]model.SystemSetting)
	for _, s := range all {
		if !s.IsPublic && isSensitiveKey(s.Key) {
			s.Value = "******"
		}
		grouped[s.Category] = append(grouped[s.Category], s)
	}
	response.OK(c, grouped)
}

// GetSetting returns a single setting
// GET /api/v1/settings/:category/:key
func (h *SystemSettingHandler) GetSetting(c *gin.Context) {
	s, err := h.repo.Get(c.Request.Context(), c.Param("category"), c.Param("key"))
	if err != nil {
		response.NotFound(c, "setting not found")
		return
	}
	response.OK(c, s)
}

// UpdateSetting updates a setting
// PUT /api/v1/settings/:category/:key
func (h *SystemSettingHandler) UpdateSetting(c *gin.Context) {
	var req model.SystemSettingUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	setting := &model.SystemSetting{
		Category: c.Param("category"),
		Key:      c.Param("key"),
		Value:    req.Value,
		Remark:   req.Remark,
	}
	if err := h.repo.Upsert(c.Request.Context(), setting); err != nil {
		response.InternalError(c, "failed to update setting")
		return
	}
	response.OKMsg(c, "updated")
}

// BatchUpdateSettings updates multiple settings at once
// POST /api/v1/settings/batch
func (h *SystemSettingHandler) BatchUpdateSettings(c *gin.Context) {
	var settings []model.SystemSetting
	if err := c.ShouldBindJSON(&settings); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.repo.BatchUpsert(c.Request.Context(), settings); err != nil {
		response.InternalError(c, "failed to update settings")
		return
	}
	response.OKMsg(c, "settings updated")
}

// GetSMTPConfig returns the SMTP configuration
// GET /api/v1/settings/smtp
func (h *SystemSettingHandler) GetSMTPConfig(c *gin.Context) {
	h.getCategorySettings(c, "smtp")
}

// UpdateSMTPConfig updates SMTP settings
// POST /api/v1/settings/smtp
func (h *SystemSettingHandler) UpdateSMTPConfig(c *gin.Context) {
	var cfg model.SMTPConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	settings := smtpToSettings(cfg)
	if err := h.repo.BatchUpsert(c.Request.Context(), settings); err != nil {
		response.InternalError(c, "failed to update SMTP config")
		return
	}
	response.OKMsg(c, "SMTP config updated")
}

// GetStorageConfig returns storage configuration
// GET /api/v1/settings/storage
func (h *SystemSettingHandler) GetStorageConfig(c *gin.Context) {
	h.getCategorySettings(c, "storage")
}

// UpdateStorageConfig updates storage settings
// POST /api/v1/settings/storage
func (h *SystemSettingHandler) UpdateStorageConfig(c *gin.Context) {
	var cfg model.StorageConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	settings := storageToSettings(cfg)
	if err := h.repo.BatchUpsert(c.Request.Context(), settings); err != nil {
		response.InternalError(c, "failed to update storage config")
		return
	}
	response.OKMsg(c, "storage config updated")
}

// GetSecurityConfig returns security settings
// GET /api/v1/settings/security
func (h *SystemSettingHandler) GetSecurityConfig(c *gin.Context) {
	h.getCategorySettings(c, "security")
}

// UpdateSecurityConfig updates security settings
// POST /api/v1/settings/security
func (h *SystemSettingHandler) UpdateSecurityConfig(c *gin.Context) {
	var cfg model.SecurityConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	settings := securityToSettings(cfg)
	if err := h.repo.BatchUpsert(c.Request.Context(), settings); err != nil {
		response.InternalError(c, "failed to update security config")
		return
	}
	response.OKMsg(c, "security config updated")
}

// GetSystemInfo returns system runtime information
// GET /api/v1/settings/system-info
func (h *SystemSettingHandler) GetSystemInfo(c *gin.Context) {
	info := &model.SystemInfo{
		Version:   "2.1.0",
		StartTime: "", // set at app startup
	}
	response.OK(c, info)
}

func (h *SystemSettingHandler) getCategorySettings(c *gin.Context, category string) {
	list, err := h.repo.ListByCategory(c.Request.Context(), category)
	if err != nil {
		response.InternalError(c, "failed to load settings")
		return
	}
	// Mask sensitive values for non-system internal settings
	masked := make([]model.SystemSetting, len(list))
	for i, s := range list {
		if !s.IsPublic && isSensitiveKey(s.Key) {
			s.Value = "******"
		}
		masked[i] = s
	}
	response.OK(c, masked)
}

// isSensitiveKey identifies setting keys that should be masked in responses
func isSensitiveKey(key string) bool {
	sensitiveKeys := map[string]bool{
		"password": true, "secret_key": true, "access_key": true,
		"smtp_password": true, "api_key": true, "token": true,
	}
	return sensitiveKeys[key]
}

func smtpToSettings(cfg model.SMTPConfig) []model.SystemSetting {
	return []model.SystemSetting{
		{Category: "smtp", Key: "host", Value: cfg.Host, ValueType: "string", Remark: "SMTP服务器地址"},
		{Category: "smtp", Key: "port", Value: strconv.Itoa(cfg.Port), ValueType: "int", Remark: "SMTP端口"},
		{Category: "smtp", Key: "user", Value: cfg.User, ValueType: "string", Remark: "SMTP用户名"},
		{Category: "smtp", Key: "password", Value: cfg.Password, ValueType: "string", Remark: "SMTP密码", IsPublic: false},
		{Category: "smtp", Key: "from_name", Value: cfg.FromName, ValueType: "string", Remark: "发件人名称"},
		{Category: "smtp", Key: "from_addr", Value: cfg.FromAddr, ValueType: "string", Remark: "发件人邮箱"},
		{Category: "smtp", Key: "use_tls", Value: strconv.FormatBool(cfg.UseTLS), ValueType: "bool", Remark: "启用TLS"},
	}
}

func storageToSettings(cfg model.StorageConfig) []model.SystemSetting {
	return []model.SystemSetting{
		{Category: "storage", Key: "provider", Value: cfg.Provider, ValueType: "string", Remark: "存储类型"},
		{Category: "storage", Key: "endpoint", Value: cfg.Endpoint, ValueType: "string", Remark: "存储端点"},
		{Category: "storage", Key: "bucket", Value: cfg.Bucket, ValueType: "string", Remark: "存储桶"},
		{Category: "storage", Key: "access_key", Value: cfg.AK, ValueType: "string", Remark: "Access Key", IsPublic: false},
		{Category: "storage", Key: "secret_key", Value: cfg.SK, ValueType: "string", Remark: "Secret Key", IsPublic: false},
		{Category: "storage", Key: "region", Value: cfg.Region, ValueType: "string", Remark: "区域"},
		{Category: "storage", Key: "path_prefix", Value: cfg.PathPrefix, ValueType: "string", Remark: "路径前缀"},
	}
}

func securityToSettings(cfg model.SecurityConfig) []model.SystemSetting {
	return []model.SystemSetting{
		{Category: "security", Key: "password_min_length", Value: strconv.Itoa(cfg.PasswordMinLength), ValueType: "int", Remark: "密码最小长度"},
		{Category: "security", Key: "password_require_upper", Value: strconv.FormatBool(cfg.PasswordRequireUpper), ValueType: "bool", Remark: "密码需大写字母"},
		{Category: "security", Key: "password_require_number", Value: strconv.FormatBool(cfg.PasswordRequireNumber), ValueType: "bool", Remark: "密码需数字"},
		{Category: "security", Key: "password_require_special", Value: strconv.FormatBool(cfg.PasswordRequireSpecial), ValueType: "bool", Remark: "密码需特殊字符"},
		{Category: "security", Key: "login_max_attempts", Value: strconv.Itoa(cfg.LoginMaxAttempts), ValueType: "int", Remark: "登录最大尝试次数"},
		{Category: "security", Key: "login_lock_duration", Value: strconv.Itoa(cfg.LoginLockDuration), ValueType: "int", Remark: "锁定时长(分钟)"},
		{Category: "security", Key: "session_timeout", Value: strconv.Itoa(cfg.SessionTimeout), ValueType: "int", Remark: "会话超时(小时)"},
		{Category: "security", Key: "ip_whitelist", Value: cfg.IPWhitelist, ValueType: "string", Remark: "IP白名单"},
		{Category: "security", Key: "enable_2fa", Value: strconv.FormatBool(cfg.Enable2FA), ValueType: "bool", Remark: "启用双因素认证"},
	}
}
