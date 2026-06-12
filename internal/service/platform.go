package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ═══ Platform SDK Connectors ═══
// Production-grade connectors for Douyin, Kuaishou, Taobao Live
// Each implements PlatformConnector from engine.go

// ─── Base HTTP Client ───

type PlatformHTTPClient struct {
	Client    *http.Client
	BaseURL   string
	AppKey    string
	AppSecret string
	Token     string
	mu        sync.RWMutex
}

func newPlatformHTTP(baseURL, appKey, appSecret string) *PlatformHTTPClient {
	return &PlatformHTTPClient{
		Client:    &http.Client{Timeout: 30 * time.Second},
		BaseURL:   baseURL,
		AppKey:    appKey,
		AppSecret: appSecret,
	}
}

func (c *PlatformHTTPClient) SetToken(token string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Token = token
}

func (c *PlatformHTTPClient) GetToken() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Token
}

func (c *PlatformHTTPClient) doRequest(ctx context.Context, method, path string, params map[string]string, body io.Reader) ([]byte, error) {
	u, _ := url.Parse(c.BaseURL + path)
	if method == http.MethodGet && len(params) > 0 {
		q := u.Query()
		for k, v := range params {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if token := c.GetToken(); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return data, fmt.Errorf("API error %d: %s", resp.StatusCode, string(data))
	}
	return data, nil
}

// ═══ Douyin (TikTok China) Connector ═══

type DouyinConnector struct {
	client *PlatformHTTPClient
}

func NewDouyinConnector(appKey, appSecret string) *DouyinConnector {
	return &DouyinConnector{
		client: newPlatformHTTP("https://open.douyin.com", appKey, appSecret),
	}
}

func (d *DouyinConnector) Connect(config string) error {
	var cfg struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal([]byte(config), &cfg); err != nil {
		return fmt.Errorf("parse douyin config: %w", err)
	}
	if cfg.AccessToken == "" {
		// Try OAuth2 refresh
		return fmt.Errorf("douyin access_token required")
	}
	d.client.SetToken(cfg.AccessToken)
	return nil
}

func (d *DouyinConnector) FetchLiveRooms(ctx context.Context) ([]json.RawMessage, error) {
	data, err := d.client.doRequest(ctx, http.MethodGet, "/api/douyin/lives/", map[string]string{
		"page":  "1",
		"count": "50",
	}, nil)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Data struct {
			List []json.RawMessage `json:"list"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parse douyin rooms: %w", err)
	}
	return resp.Data.List, nil
}

func (d *DouyinConnector) FetchOrders(ctx context.Context, since time.Time) ([]json.RawMessage, error) {
	data, err := d.client.doRequest(ctx, http.MethodGet, "/api/douyin/order/list/", map[string]string{
		"start_time": strconv.FormatInt(since.Unix(), 10),
		"end_time":   strconv.FormatInt(time.Now().Unix(), 10),
		"page":       "1",
		"count":      "100",
	}, nil)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Data struct {
			List []json.RawMessage `json:"list"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parse douyin orders: %w", err)
	}
	return resp.Data.List, nil
}

func (d *DouyinConnector) FetchProducts(ctx context.Context) ([]json.RawMessage, error) {
	data, err := d.client.doRequest(ctx, http.MethodGet, "/api/douyin/product/list/", map[string]string{
		"page":  "1",
		"count": "100",
	}, nil)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Data struct {
			List []json.RawMessage `json:"list"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parse douyin products: %w", err)
	}
	return resp.Data.List, nil
}

// FetchLiveMetrics retrieves real-time metrics for a live room
func (d *DouyinConnector) FetchLiveMetrics(ctx context.Context, roomID string) (map[string]interface{}, error) {
	data, err := d.client.doRequest(ctx, http.MethodGet, "/api/douyin/lives/"+roomID+"/metrics/", nil, nil)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// RefreshToken handles OAuth2 token refresh
func (d *DouyinConnector) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	data, err := d.client.doRequest(ctx, http.MethodPost, "/oauth/refresh_token/", map[string]string{
		"client_key":    d.client.AppKey,
		"refresh_token": refreshToken,
		"grant_type":    "refresh_token",
	}, strings.NewReader("{}"))
	if err != nil {
		return "", err
	}
	var resp struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", err
	}
	d.client.SetToken(resp.Data.AccessToken)
	return resp.Data.AccessToken, nil
}

// ═══ Kuaishou Connector ═══

type KuaishouConnector struct {
	client *PlatformHTTPClient
}

func NewKuaishouConnector(appKey, appSecret string) *KuaishouConnector {
	return &KuaishouConnector{
		client: newPlatformHTTP("https://open.kuaishou.com", appKey, appSecret),
	}
}

func (k *KuaishouConnector) Connect(config string) error {
	var cfg struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal([]byte(config), &cfg); err != nil {
		return fmt.Errorf("parse kuaishou config: %w", err)
	}
	if cfg.AccessToken == "" {
		return fmt.Errorf("kuaishou access_token required")
	}
	k.client.SetToken(cfg.AccessToken)
	return nil
}

func (k *KuaishouConnector) FetchLiveRooms(ctx context.Context) ([]json.RawMessage, error) {
	data, err := k.client.doRequest(ctx, http.MethodGet, "/openapi/live/list", map[string]string{
		"page_size": "50",
	}, nil)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Result struct {
			LiveStreamList []json.RawMessage `json:"liveStreamList"`
		} `json:"result"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parse kuaishou rooms: %w", err)
	}
	return resp.Result.LiveStreamList, nil
}

func (k *KuaishouConnector) FetchOrders(ctx context.Context, since time.Time) ([]json.RawMessage, error) {
	data, err := k.client.doRequest(ctx, http.MethodGet, "/openapi/order/list", map[string]string{
		"begin_time": since.Format("2006-01-02 15:04:05"),
		"end_time":   time.Now().Format("2006-01-02 15:04:05"),
		"page_size":  "100",
	}, nil)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Result struct {
			OrderList []json.RawMessage `json:"orderList"`
		} `json:"result"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parse kuaishou orders: %w", err)
	}
	return resp.Result.OrderList, nil
}

func (k *KuaishouConnector) FetchProducts(ctx context.Context) ([]json.RawMessage, error) {
	data, err := k.client.doRequest(ctx, http.MethodGet, "/openapi/goods/list", map[string]string{
		"page_size": "100",
	}, nil)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Result struct {
			GoodsList []json.RawMessage `json:"goodsList"`
		} `json:"result"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parse kuaishou products: %w", err)
	}
	return resp.Result.GoodsList, nil
}

// FetchLiveStats retrieves real-time stats for a live stream
func (k *KuaishouConnector) FetchLiveStats(ctx context.Context, liveStreamID string) (map[string]interface{}, error) {
	data, err := k.client.doRequest(ctx, http.MethodGet, "/openapi/live/stats", map[string]string{
		"live_stream_id": liveStreamID,
	}, nil)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ═══ Taobao Live Connector ═══

type TaobaoConnector struct {
	client *PlatformHTTPClient
}

func NewTaobaoConnector(appKey, appSecret string) *TaobaoConnector {
	return &TaobaoConnector{
		client: newPlatformHTTP("https://eco.taobao.com/router/rest", appKey, appSecret),
	}
}

func (t *TaobaoConnector) Connect(config string) error {
	var cfg struct {
		Session string `json:"session"`
	}
	if err := json.Unmarshal([]byte(config), &cfg); err != nil {
		return fmt.Errorf("parse taobao config: %w", err)
	}
	if cfg.Session == "" {
		return fmt.Errorf("taobao session key required")
	}
	t.client.SetToken(cfg.Session)
	return nil
}

func (t *TaobaoConnector) FetchLiveRooms(ctx context.Context) ([]json.RawMessage, error) {
	data, err := t.client.doRequest(ctx, http.MethodGet, "/", map[string]string{
		"method":     "taobao.live.room.get",
		"app_key":    t.client.AppKey,
		"session":    t.client.GetToken(),
		"page_no":    "1",
		"page_size":  "50",
		"timestamp":  time.Now().Format("2006-01-02 15:04:05"),
		"format":     "json",
		"v":          "2.0",
		"sign_method": "hmac",
	}, nil)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Rsp struct {
			Results struct {
				Rooms []json.RawMessage `json:"room"`
			} `json:"results"`
		} `json:"live_room_get_response"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parse taobao rooms: %w", err)
	}
	return resp.Rsp.Results.Rooms, nil
}

func (t *TaobaoConnector) FetchOrders(ctx context.Context, since time.Time) ([]json.RawMessage, error) {
	data, err := t.client.doRequest(ctx, http.MethodGet, "/", map[string]string{
		"method":      "taobao.trades.sold.get",
		"app_key":     t.client.AppKey,
		"session":     t.client.GetToken(),
		"start_created": since.Format("2006-01-02 15:04:05"),
		"end_created":   time.Now().Format("2006-01-02 15:04:05"),
		"page_no":     "1",
		"page_size":   "100",
		"format":      "json",
		"v":           "2.0",
	}, nil)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Rsp struct {
			Trades struct {
				Trade []json.RawMessage `json:"trade"`
			} `json:"trades"`
		} `json:"trades_sold_get_response"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parse taobao orders: %w", err)
	}
	return resp.Rsp.Trades.Trade, nil
}

func (t *TaobaoConnector) FetchProducts(ctx context.Context) ([]json.RawMessage, error) {
	data, err := t.client.doRequest(ctx, http.MethodGet, "/", map[string]string{
		"method":     "taobao.items.onsale.get",
		"app_key":    t.client.AppKey,
		"session":    t.client.GetToken(),
		"page_no":    "1",
		"page_size":  "100",
		"format":     "json",
		"v":          "2.0",
	}, nil)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Rsp struct {
			Items struct {
				Item []json.RawMessage `json:"item"`
			} `json:"items"`
		} `json:"items_onsale_get_response"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parse taobao products: %w", err)
	}
	return resp.Rsp.Items.Item, nil
}

// FetchLivePlayback retrieves playback data for a completed live stream
func (t *TaobaoConnector) FetchLivePlayback(ctx context.Context, roomID string) (map[string]interface{}, error) {
	data, err := t.client.doRequest(ctx, http.MethodGet, "/", map[string]string{
		"method":   "taobao.live.playback.get",
		"app_key":  t.client.AppKey,
		"session":  t.client.GetToken(),
		"room_id":  roomID,
		"format":   "json",
		"v":        "2.0",
	}, nil)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ═══ Platform Connector Factory ═══

var platformFactories = map[string]func(appKey, appSecret string) PlatformConnector{
	"douyin":     func(k, s string) PlatformConnector { return NewDouyinConnector(k, s) },
	"kuaishou":   func(k, s string) PlatformConnector { return NewKuaishouConnector(k, s) },
	"taobao_live": func(k, s string) PlatformConnector { return NewTaobaoConnector(k, s) },
}

// CreateConnector creates a platform connector by name
func CreateConnector(platform, appKey, appSecret string) (PlatformConnector, error) {
	factory, ok := platformFactories[platform]
	if !ok {
		return nil, fmt.Errorf("unsupported platform: %s (supported: douyin, kuaishou, taobao_live)", platform)
	}
	return factory(appKey, appSecret), nil
}

// ListSupportedPlatforms returns all supported platform names
func ListSupportedPlatforms() []string {
	platforms := make([]string, 0, len(platformFactories))
	for k := range platformFactories {
		platforms = append(platforms, k)
	}
	return platforms
}
