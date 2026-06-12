package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/CodingFervor/live-commerce-bi/internal/database"
	"github.com/CodingFervor/live-commerce-bi/internal/model"
	"github.com/CodingFervor/live-commerce-bi/pkg/logger"
)

// ═══ Multi-Model AI Service ═══
// Unified interface for OpenAI, Qwen (通义千问), Zhipu GLM, Baidu Wenxin,
// DeepSeek, Moonshot Kimi, iFlytek Spark, and local Ollama

// AIProviderConfig holds runtime provider settings
type AIProviderConfig struct {
	Endpoint    string
	APIKey      string
	Model       string
	MaxTokens   int
	Temperature float64
	TopP        float64
	ProxyURL    string
}

// ChatCompletionRequest is the unified request format
type ChatCompletionRequest struct {
	Model       string        `json:"model"`
	Messages    []AIMessage   `json:"messages"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
	TopP        float64       `json:"top_p,omitempty"`
	Stream      bool          `json:"stream"`
}

// AIMessage mirrors model.AIMessage for internal use
type AIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionResponse is the unified response
type ChatCompletionResponse struct {
	ID      string `json:"id"`
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// AIService provides AI capabilities
type AIService struct {
	httpClient *http.Client
}

func NewAIService() *AIService {
	return &AIService{
		httpClient: &http.Client{Timeout: 120 * time.Second},
	}
}

// ─── Core Chat Completion ───

func (s *AIService) Chat(ctx context.Context, config *model.AIConfig, messages []model.AIMessage) (*model.AITestResponse, error) {
	start := time.Now()

	// Build provider-specific request
	providerCfg := s.buildProviderConfig(config)
	reqMessages := make([]AIMessage, len(messages))
	for i, m := range messages {
		reqMessages[i] = AIMessage{Role: m.Role, Content: m.Content}
	}

	reqBody := &ChatCompletionRequest{
		Model:       providerCfg.Model,
		Messages:    reqMessages,
		MaxTokens:   providerCfg.MaxTokens,
		Temperature: providerCfg.Temperature,
		TopP:        providerCfg.TopP,
		Stream:      false,
	}

	endpoint := s.buildEndpoint(config)
	resp, err := s.doRequest(ctx, endpoint, providerCfg.APIKey, reqBody, config.ProxyURL)
	if err != nil {
		return &model.AITestResponse{
			Success:   false,
			Error:     err.Error(),
			LatencyMs: time.Since(start).Milliseconds(),
		}, err
	}

	reply := ""
	tokensUsed := 0
	if len(resp.Choices) > 0 {
		reply = resp.Choices[0].Message.Content
	}
	if resp.Usage.TotalTokens > 0 {
		tokensUsed = resp.Usage.TotalTokens
	}

	return &model.AITestResponse{
		Success:    true,
		Reply:      reply,
		TokensUsed: tokensUsed,
		LatencyMs:  time.Since(start).Milliseconds(),
	}, nil
}

// TestConnection tests if an AI provider is reachable
func (s *AIService) TestConnection(ctx context.Context, config *model.AIConfig) (*model.AITestResponse, error) {
	testMessages := []model.AIMessage{
		{Role: "user", Content: "请回复\"连接成功\""},
	}
	return s.Chat(ctx, config, testMessages)
}

// ─── BI Intelligence Features ───

// SmartQuery converts natural language to data insights
func (s *AIService) SmartQuery(ctx context.Context, config *model.AIConfig, query string, dataCtx *model.AIContext) (*model.AIChatResponse, error) {
	systemPrompt := s.buildBISystemPrompt(dataCtx)
	messages := []model.AIMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: query},
	}

	resp, err := s.Chat(ctx, config, messages)
	if err != nil {
		return nil, err
	}

	return &model.AIChatResponse{
		Reply:      resp.Reply,
		TokensUsed: resp.TokensUsed,
	}, nil
}

// GenerateInsights generates AI insights for a given analytics topic
func (s *AIService) GenerateInsights(ctx context.Context, config *model.AIConfig, req *model.AIInsightRequest, analyticsData interface{}) (*model.AIInsightResponse, error) {
	dataJSON, _ := json.Marshal(analyticsData)

	systemPrompt := fmt.Sprintf(`你是一个直播电商BI数据分析师。请基于以下%s数据，生成深度分析和可操作建议。
要求：
1. 提供简洁的数据摘要（100字内）
2. 列出3-5个关键洞察
3. 提出3-5个可操作建议
4. 回复使用纯JSON格式: {"summary":"...","insights":["..."],"recommendations":["..."]}

数据范围: %s ~ %s, 平台: %s
数据内容:`, req.Topic, req.StartDate, req.EndDate, req.Platform)

	messages := []model.AIMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: string(dataJSON)},
	}

	resp, err := s.Chat(ctx, config, messages)
	if err != nil {
		return nil, err
	}

	insightResp := &model.AIInsightResponse{
		Topic:      req.Topic,
		TokensUsed: resp.TokensUsed,
	}

	// Try to parse structured JSON response
	var structured struct {
		Summary         string   `json:"summary"`
		Insights        []string `json:"insights"`
		Recommendations []string `json:"recommendations"`
	}
	content := resp.Reply
	// Extract JSON from possible markdown code block
	if idx := strings.Index(content, "{"); idx >= 0 {
		content = content[idx:]
		if endIdx := strings.LastIndex(content, "}"); endIdx > 0 {
			content = content[:endIdx+1]
		}
	}
	if err := json.Unmarshal([]byte(content), &structured); err == nil {
		insightResp.Summary = structured.Summary
		insightResp.Insights = structured.Insights
		insightResp.Recommendations = structured.Recommendations
	} else {
		insightResp.Summary = resp.Reply
		insightResp.Insights = []string{"AI response was not in structured format"}
	}

	return insightResp, nil
}

// GenerateReport generates an AI-powered analytical report
func (s *AIService) GenerateReport(ctx context.Context, config *model.AIConfig, reportType string, data interface{}) (string, error) {
	dataJSON, _ := json.Marshal(data)

	systemPrompt := `你是一个专业的直播电商数据分析报告撰写专家。
请根据提供的数据生成一份结构化的分析报告，包含：
1. 报告标题
2. 核心摘要
3. 详细分析（包含数据趋势、关键指标变化、异常点）
4. 结论与建议

报告应使用 Markdown 格式，专业但易于理解。`

	messages := []model.AIMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: fmt.Sprintf("报告类型: %s\n数据:\n%s", reportType, string(dataJSON))},
	}

	resp, err := s.Chat(ctx, config, messages)
	if err != nil {
		return "", err
	}
	return resp.Reply, nil
}

// ─── Provider Configuration Helpers ───

func (s *AIService) buildProviderConfig(config *model.AIConfig) *AIProviderConfig {
	cfg := &AIProviderConfig{
		APIKey:      config.APIKey,
		Model:       config.ModelName,
		MaxTokens:   config.MaxTokens,
		Temperature: config.Temperature,
		TopP:        config.TopP,
		ProxyURL:    config.ProxyURL,
	}

	if cfg.MaxTokens == 0 {
		cfg.MaxTokens = 4096
	}
	if cfg.Temperature == 0 {
		cfg.Temperature = 0.7
	}
	if cfg.TopP == 0 {
		cfg.TopP = 0.9
	}

	// Set default endpoints per provider
	if config.APIEndpoint != "" {
		cfg.Endpoint = config.APIEndpoint
	} else {
		switch model.AIProvider(config.Provider) {
		case model.AIProviderOpenAI:
			cfg.Endpoint = "https://api.openai.com/v1/chat/completions"
		case model.AIProviderQwen:
			cfg.Endpoint = "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions"
		case model.AIProviderZhipu:
			cfg.Endpoint = "https://open.bigmodel.cn/api/paas/v4/chat/completions"
		case model.AIProviderBaidu:
			cfg.Endpoint = "https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/chat/completions"
		case model.AIProviderDeepSeek:
			cfg.Endpoint = "https://api.deepseek.com/v1/chat/completions"
		case model.AIProviderMoonshot:
			cfg.Endpoint = "https://api.moonshot.cn/v1/chat/completions"
		case model.AIProviderSpark:
			cfg.Endpoint = "https://spark-api.xf-yun.com/v1/chat/completions"
		case model.AIProviderOllama:
			cfg.Endpoint = "http://localhost:11434/v1/chat/completions"
		}
	}

	return cfg
}

func (s *AIService) buildEndpoint(config *model.AIConfig) string {
	cfg := s.buildProviderConfig(config)
	return cfg.Endpoint
}

func (s *AIService) doRequest(ctx context.Context, endpoint, apiKey string, body interface{}, proxyURL string) (*ChatCompletionResponse, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("AI request failed: %w", err)
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		logger.Error("AI API error %d for endpoint %s", resp.StatusCode, endpoint)
		return nil, fmt.Errorf("AI provider returned error (status %d)", resp.StatusCode)
	}

	var result ChatCompletionResponse
	if err := json.Unmarshal(respData, &result); err != nil {
		logger.Error("Failed to parse AI response: %v", err)
		return nil, fmt.Errorf("failed to parse AI provider response")
	}

	return &result, nil
}

func (s *AIService) buildBISystemPrompt(dataCtx *model.AIContext) string {
	base := `你是一个专业的直播电商BI数据分析师助手。你可以：
1. 用自然语言解释数据趋势和指标
2. 帮助用户理解GMV、转化率、留存率等核心指标
3. 提供数据驱动的运营建议
4. 根据数据生成SQL查询建议
5. 推荐适合的数据可视化方式

当前系统数据概要：
- 数据表：orders(订单), live_rooms(直播间), streamers(主播), products(商品), viewer_metrics(观众指标)
- 支持平台：抖音、快手、淘宝直播、京东直播、拼多多直播
- 核心指标：GMV、订单量、转化率、平均客单价、观看人数、互动率、留存率`

	if dataCtx != nil {
		base += fmt.Sprintf("\n\n当前查询上下文: 数据类型=%s, 时间范围=%s, 平台=%s",
			dataCtx.DataType, dataCtx.DateRange, dataCtx.Platform)
	}

	return base
}

// ─── AI Config Repository Helpers ───

func (s *AIService) LoadConfig(ctx context.Context, configID int64) (*model.AIConfig, error) {
	var cfg model.AIConfig
	err := database.Get().QueryRow(ctx, `
		SELECT id, name, provider, api_key, api_endpoint, model_name,
			max_tokens, temperature, top_p, is_default, is_enabled,
			proxy_url, extra_config, created_by, created_at, updated_at
		FROM ai_configs WHERE id=$1`, configID).Scan(
		&cfg.ID, &cfg.Name, &cfg.Provider, &cfg.APIKey, &cfg.APIEndpoint,
		&cfg.ModelName, &cfg.MaxTokens, &cfg.Temperature, &cfg.TopP,
		&cfg.IsDefault, &cfg.IsEnabled, &cfg.ProxyURL, &cfg.ExtraConfig,
		&cfg.CreatedBy, &cfg.CreatedAt, &cfg.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (s *AIService) LoadDefaultConfig(ctx context.Context) (*model.AIConfig, error) {
	var cfg model.AIConfig
	err := database.Get().QueryRow(ctx, `
		SELECT id, name, provider, api_key, api_endpoint, model_name,
			max_tokens, temperature, top_p, is_default, is_enabled,
			proxy_url, extra_config, created_by, created_at, updated_at
		FROM ai_configs WHERE is_default=true AND is_enabled=true LIMIT 1`).Scan(
		&cfg.ID, &cfg.Name, &cfg.Provider, &cfg.APIKey, &cfg.APIEndpoint,
		&cfg.ModelName, &cfg.MaxTokens, &cfg.Temperature, &cfg.TopP,
		&cfg.IsDefault, &cfg.IsEnabled, &cfg.ProxyURL, &cfg.ExtraConfig,
		&cfg.CreatedBy, &cfg.CreatedAt, &cfg.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("no default AI config found: %w", err)
	}
	return &cfg, nil
}

func minLen(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// suppress unused
var _ = logger.Info
var _ = minLen
