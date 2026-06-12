package service

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/CodingFervor/live-commerce-bi/internal/database"
	"github.com/CodingFervor/live-commerce-bi/internal/model"
	"github.com/CodingFervor/live-commerce-bi/pkg/logger"
)

// ═══ Douyin Compass Human-Like Scraper Engine ═══
//
// 抖音电商罗盘数据采集引擎
// 核心设计原则：
// 1. 拟人化请求节奏 — 随机延迟、渐进式浏览
// 2. 浏览器指纹模拟 — 多种 User-Agent、自然请求头
// 3. 频率限制 — 每日请求数上限、冷却机制
// 4. 会话管理 — Cookie 持久化、自动续期检测
// 5. 异常恢复 — 自动退避、会话切换

// ─── Browser Fingerprint Pool ───

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.5 Safari/605.1.15",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:126.0) Gecko/20100101 Firefox/126.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:126.0) Gecko/20100101 Firefox/126.0",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36 Edg/125.0.0.0",
}

var refererPages = []string{
	"https://compass.jinritemai.com/dashboard/live",
	"https://compass.jinritemai.com/dashboard/product",
	"https://compass.jinritemai.com/dashboard/order",
	"https://compass.jinritemai.com/dashboard/anchor",
	"https://compass.jinritemai.com/",
}

// ─── Compass Scraper Engine ───

type CompassEngine struct {
	mu       sync.Mutex
	sessions map[int64]*compassSession // cached sessions
	rng      *rand.Rand
}

type compassSession struct {
	client    *http.Client
	ua        string
	sessionID int64
	shopID    string
	proxyURL  string
}

func NewCompassEngine() *CompassEngine {
	return &CompassEngine{
		sessions: make(map[int64]*compassSession),
		rng:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// ─── Human-Like Behavior Helpers ───

// humanDelay simulates natural browsing pause between actions
// Fast actions: 1-3s, Normal: 3-8s, Slow (page load): 5-15s
func (e *CompassEngine) humanDelay(minMs, maxMs int) {
	ms := minMs + e.rng.Intn(maxMs-minMs)
	// Add ±20% jitter for more natural feel
	jitter := int(float64(ms) * 0.2)
	ms += e.rng.Intn(2*jitter) - jitter
	if ms < 0 {
		ms = minMs
	}
	time.Sleep(time.Duration(ms) * time.Millisecond)
}

// shortPause simulates a brief pause (like reading a number)
func (e *CompassEngine) shortPause() {
	e.humanDelay(500, 2000)
}

// pageLoadPause simulates page load + reading time
func (e *CompassEngine) pageLoadPause() {
	e.humanDelay(3000, 8000)
}

// sectionPause simulates switching between dashboard sections
func (e *CompassEngine) sectionPause() {
	e.humanDelay(5000, 15000)
}

// randomUA returns a random user agent string
func (e *CompassEngine) randomUA() string {
	return userAgents[e.rng.Intn(len(userAgents))]
}

// randomReferer returns a realistic referer for compass navigation
func (e *CompassEngine) randomReferer() string {
	return refererPages[e.rng.Intn(len(refererPages))]
}

// shouldTakeBreak decides if we need a longer rest (every 15-30 requests)
func (e *CompassEngine) shouldTakeBreak(requestCount int) bool {
	return requestCount > 0 && requestCount%15 == 0
}

// breakPause simulates taking a break (1-3 minutes)
func (e *CompassEngine) breakPause() {
	secs := 60 + e.rng.Intn(120)
	logger.Info("[Compass] Taking a break for %d seconds...", secs)
	time.Sleep(time.Duration(secs) * time.Second)
}

// ─── Session Management ───

// createHTTPClient builds a browser-like HTTP client with cookie persistence
func (e *CompassEngine) createHTTPClient(cookieStr, proxyURL string) (*http.Client, string, error) {
	ua := e.randomUA()

	jar, _ := cookiejar.New(nil)

	// Parse cookies and add to jar
	baseURL, _ := url.Parse("https://compass.jinritemai.com")
	cookies := parseCookies(cookieStr)
	jar.SetCookies(baseURL, cookies)

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12},
		MaxIdleConns:    10,
		IdleConnTimeout: 120 * time.Second,
	}

	// Proxy support
	if proxyURL != "" {
		proxy, err := url.Parse(proxyURL)
		if err == nil {
			transport.Proxy = http.ProxyURL(proxy)
		}
	}

	client := &http.Client{
		Transport: transport,
		Jar:       jar,
		Timeout:   30 * time.Second,
	}

	return client, ua, nil
}

func parseCookies(cookieStr string) []*http.Cookie {
	var cookies []*http.Cookie
	pairs := strings.Split(cookieStr, ";")
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			cookies = append(cookies, &http.Cookie{
				Name:  strings.TrimSpace(parts[0]),
				Value: strings.TrimSpace(parts[1]),
			})
		}
	}
	return cookies
}

// getSession loads or creates a cached HTTP session for a compass account
func (e *CompassEngine) getSession(sessionID int64) (*compassSession, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if s, ok := e.sessions[sessionID]; ok {
		return s, nil
	}

	// Load session from database
	sess, err := e.loadSessionFromDB(sessionID)
	if err != nil {
		return nil, fmt.Errorf("session %d not found: %w", sessionID, err)
	}

	e.sessions[sessionID] = sess
	return sess, nil
}

func (e *CompassEngine) loadSessionFromDB(sessionID int64) (*compassSession, error) {
	var cookie, shopID, proxyURL, ua string
	err := database.Get().QueryRow(context.Background(),
		`SELECT cookie, shop_id, proxy_url, user_agent FROM compass_sessions WHERE id=$1 AND status='active'`,
		sessionID).Scan(&cookie, &shopID, &proxyURL, &ua)
	if err != nil {
		return nil, err
	}

	client, resolvedUA, err := e.createHTTPClient(cookie, proxyURL)
	if err != nil {
		return nil, err
	}

	if ua != "" {
		resolvedUA = ua
	}

	return &compassSession{
		client:    client,
		ua:        resolvedUA,
		sessionID: sessionID,
		shopID:    shopID,
		proxyURL:  proxyURL,
	}, nil
}

// InvalidateSession removes cached session (e.g., after cookie update)
func (e *CompassEngine) InvalidateSession(sessionID int64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.sessions, sessionID)
}

// ─── Core HTTP Request with Human-Like Behavior ───

type compassResponse struct {
	Data   json.RawMessage `json:"data"`
	Code   int             `json:"code"`
	Status int             `json:"status"`
	Msg    string          `json:"msg"`
}

func (e *CompassEngine) doCompassRequest(ctx context.Context, sess *compassSession, apiPath string, query url.Values) ([]byte, error) {
	reqURL := fmt.Sprintf("https://compass.jinritemai.com%s", apiPath)
	if len(query) > 0 {
		reqURL += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Browser-like headers (order matters for fingerprint detection)
	req.Header.Set("User-Agent", sess.ua)
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Referer", e.randomReferer())
	req.Header.Set("Origin", "https://compass.jinritemai.com")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Sec-Ch-Ua", `"Chromium";v="125", "Google Chrome";v="125"`)
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", `"Windows"`)
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")

	resp, err := sess.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	// Check for ban/captcha/redirect indicators
	if resp.StatusCode == 403 {
		return nil, fmt.Errorf("access denied (403) - session may be banned or cookie expired")
	}
	if resp.StatusCode == 302 {
		return nil, fmt.Errorf("redirected (302) - session expired, cookie needs refresh")
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body[:min200(len(body), 500)]))
	}

	// Check for login redirect in response body
	bodyStr := string(body)
	if strings.Contains(bodyStr, "login") && strings.Contains(bodyStr, "passport") {
		return nil, fmt.Errorf("session expired - redirected to login page")
	}

	// Increment daily request counter
	e.incrementRequestCount(sess.sessionID)

	return body, nil
}

func min200(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ─── Data Fetching Methods (with human-like pacing) ───

// FetchLiveOverview fetches live streaming overview data from compass
// API: /api/dashboard/live/overview
func (e *CompassEngine) FetchLiveOverview(ctx context.Context, sessionID int64, startDate, endDate string) ([]model.CompassLiveOverview, error) {
	sess, err := e.getSession(sessionID)
	if err != nil {
		return nil, err
	}

	// Check rate limits before starting
	if !e.checkRateLimit(sessionID) {
		return nil, fmt.Errorf("daily request limit reached, try again tomorrow")
	}

	logger.Info("[Compass] Fetching live overview: %s ~ %s (session=%d)", startDate, endDate, sessionID)

	// Simulate: user opens dashboard, takes a moment to load
	e.pageLoadPause()

	params := url.Values{}
	params.Set("start_date", startDate)
	params.Set("end_date", endDate)
	params.Set("shop_id", sess.shopID)

	body, err := e.doCompassRequest(ctx, sess, "/api/dashboard/live/overview", params)
	if err != nil {
		return nil, e.handleFetchError(sessionID, err)
	}

	// Simulate: user reads the page
	e.shortPause()

	var resp struct {
		Code int                       `json:"code"`
		Data struct {
			List []model.CompassLiveOverview `json:"list"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse live overview: %w", err)
	}

	return resp.Data.List, nil
}

// FetchLiveDetail fetches individual live room details
// Simulates: user clicks into a specific live room to view details
func (e *CompassEngine) FetchLiveDetail(ctx context.Context, sessionID int64, roomID string) (*model.CompassLiveDetail, error) {
	sess, err := e.getSession(sessionID)
	if err != nil {
		return nil, err
	}

	if !e.checkRateLimit(sessionID) {
		return nil, fmt.Errorf("daily request limit reached")
	}

	logger.Info("[Compass] Fetching live detail: room=%s (session=%d)", roomID, sessionID)

	// Simulate: user clicks into a specific live room
	e.humanDelay(2000, 5000)

	params := url.Values{}
	params.Set("room_id", roomID)
	params.Set("shop_id", sess.shopID)

	body, err := e.doCompassRequest(ctx, sess, "/api/dashboard/live/detail", params)
	if err != nil {
		return nil, e.handleFetchError(sessionID, err)
	}

	// Simulate: user spends time reading the detail page
	e.pageLoadPause()

	var resp struct {
		Code int                    `json:"code"`
		Data model.CompassLiveDetail `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse live detail: %w", err)
	}

	return &resp.Data, nil
}

// FetchProductList fetches product listing data
// Simulates: user navigates to product page and browses
func (e *CompassEngine) FetchProductList(ctx context.Context, sessionID int64, page, pageSize int, category string) ([]model.CompassProductBrief, error) {
	sess, err := e.getSession(sessionID)
	if err != nil {
		return nil, err
	}

	if !e.checkRateLimit(sessionID) {
		return nil, fmt.Errorf("daily request limit reached")
	}

	logger.Info("[Compass] Fetching product list: page=%d (session=%d)", page, sessionID)

	// Simulate: user navigates to product section
	e.sectionPause()

	params := url.Values{}
	params.Set("page", fmt.Sprintf("%d", page))
	params.Set("page_size", fmt.Sprintf("%d", pageSize))
	params.Set("shop_id", sess.shopID)
	if category != "" {
		params.Set("category", category)
	}

	body, err := e.doCompassRequest(ctx, sess, "/api/dashboard/product/list", params)
	if err != nil {
		return nil, e.handleFetchError(sessionID, err)
	}

	// Simulate: user scrolls through product list
	e.humanDelay(2000, 5000)

	var resp struct {
		Code int `json:"code"`
		Data struct {
			List []model.CompassProductBrief `json:"list"`
			Total int                        `json:"total"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse product list: %w", err)
	}

	return resp.Data.List, nil
}

// FetchProductDetail fetches detailed product analytics
// Simulates: user clicks into a product to view analytics
func (e *CompassEngine) FetchProductDetail(ctx context.Context, sessionID int64, productID string) (*model.CompassProductDetail, error) {
	sess, err := e.getSession(sessionID)
	if err != nil {
		return nil, err
	}

	if !e.checkRateLimit(sessionID) {
		return nil, fmt.Errorf("daily request limit reached")
	}

	logger.Info("[Compass] Fetching product detail: product=%s (session=%d)", productID, sessionID)

	// Simulate: user clicks into product detail
	e.humanDelay(2000, 4000)

	params := url.Values{}
	params.Set("product_id", productID)
	params.Set("shop_id", sess.shopID)

	body, err := e.doCompassRequest(ctx, sess, "/api/dashboard/product/detail", params)
	if err != nil {
		return nil, e.handleFetchError(sessionID, err)
	}

	e.shortPause()

	var resp struct {
		Code int                      `json:"code"`
		Data model.CompassProductDetail `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse product detail: %w", err)
	}

	return &resp.Data, nil
}

// FetchOrderList fetches order data from compass
// Simulates: user browses order list page by page
func (e *CompassEngine) FetchOrderList(ctx context.Context, sessionID int64, startDate, endDate string, page, pageSize int) ([]model.CompassOrderItem, error) {
	sess, err := e.getSession(sessionID)
	if err != nil {
		return nil, err
	}

	if !e.checkRateLimit(sessionID) {
		return nil, fmt.Errorf("daily request limit reached")
	}

	logger.Info("[Compass] Fetching orders: %s ~ %s page=%d (session=%d)", startDate, endDate, page, sessionID)

	// Simulate: user navigates to order section
	e.sectionPause()

	params := url.Values{}
	params.Set("start_date", startDate)
	params.Set("end_date", endDate)
	params.Set("page", fmt.Sprintf("%d", page))
	params.Set("page_size", fmt.Sprintf("%d", pageSize))
	params.Set("shop_id", sess.shopID)

	body, err := e.doCompassRequest(ctx, sess, "/api/dashboard/order/list", params)
	if err != nil {
		return nil, e.handleFetchError(sessionID, err)
	}

	// Simulate: user reads through the order list
	e.humanDelay(3000, 7000)

	var resp struct {
		Code int `json:"code"`
		Data struct {
			List []model.CompassOrderItem `json:"list"`
			Total int                     `json:"total"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse order list: %w", err)
	}

	return resp.Data.List, nil
}

// FetchStreamerRank fetches streamer ranking data
func (e *CompassEngine) FetchStreamerRank(ctx context.Context, sessionID int64, startDate, endDate string) ([]model.CompassStreamerRank, error) {
	sess, err := e.getSession(sessionID)
	if err != nil {
		return nil, err
	}

	if !e.checkRateLimit(sessionID) {
		return nil, fmt.Errorf("daily request limit reached")
	}

	logger.Info("[Compass] Fetching streamer rank: %s ~ %s (session=%d)", startDate, endDate, sessionID)

	e.sectionPause()

	params := url.Values{}
	params.Set("start_date", startDate)
	params.Set("end_date", endDate)
	params.Set("shop_id", sess.shopID)

	body, err := e.doCompassRequest(ctx, sess, "/api/dashboard/anchor/rank", params)
	if err != nil {
		return nil, e.handleFetchError(sessionID, err)
	}

	e.shortPause()

	var resp struct {
		Code int `json:"code"`
		Data struct {
			List []model.CompassStreamerRank `json:"list"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse streamer rank: %w", err)
	}

	return resp.Data.List, nil
}

// FetchFunnelAnalysis fetches conversion funnel data
func (e *CompassEngine) FetchFunnelAnalysis(ctx context.Context, sessionID int64, startDate, endDate string) ([]model.CompassFunnelData, error) {
	sess, err := e.getSession(sessionID)
	if err != nil {
		return nil, err
	}

	if !e.checkRateLimit(sessionID) {
		return nil, fmt.Errorf("daily request limit reached")
	}

	logger.Info("[Compass] Fetching funnel analysis: %s ~ %s (session=%d)", startDate, endDate, sessionID)

	e.pageLoadPause()

	params := url.Values{}
	params.Set("start_date", startDate)
	params.Set("end_date", endDate)
	params.Set("shop_id", sess.shopID)

	body, err := e.doCompassRequest(ctx, sess, "/api/dashboard/funnel/analysis", params)
	if err != nil {
		return nil, e.handleFetchError(sessionID, err)
	}

	e.shortPause()

	var resp struct {
		Code int `json:"code"`
		Data struct {
			List []model.CompassFunnelData `json:"list"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse funnel data: %w", err)
	}

	return resp.Data.List, nil
}

// ─── Full Sync Pipeline ───
// Runs a complete data collection cycle with human-like pacing

type CompassSyncResult struct {
	TotalRequests int                    `json:"total_requests"`
	Duration      string                 `json:"duration"`
	LiveOverview  []model.CompassLiveOverview  `json:"live_overview"`
	Products      []model.CompassProductBrief  `json:"products"`
	Orders        []model.CompassOrderItem     `json:"orders"`
	Streamers     []model.CompassStreamerRank  `json:"streamers"`
	Funnel        []model.CompassFunnelData    `json:"funnel"`
	Errors        []string               `json:"errors,omitempty"`
}

// RunFullSync executes a complete sync cycle with pacing
func (e *CompassEngine) RunFullSync(ctx context.Context, sessionID int64, startDate, endDate string) (*CompassSyncResult, error) {
	start := time.Now()
	result := &CompassSyncResult{}
	requestCount := 0

	// Step 1: Live overview
	logger.Info("[Compass] === Starting full sync (session=%d) ===", sessionID)

	lives, err := e.FetchLiveOverview(ctx, sessionID, startDate, endDate)
	requestCount++
	if e.shouldTakeBreak(requestCount) {
		e.breakPause()
	}
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("live_overview: %v", err))
		logger.Warn("[Compass] Live overview failed: %v", err)
	} else {
		result.LiveOverview = lives
		logger.Info("[Compass] Got %d live overview records", len(lives))
	}

	// Step 2: Product list (browse through pages like a real user)
	productPage := 1
	for {
		products, err := e.FetchProductList(ctx, sessionID, productPage, 20, "")
		requestCount++
		if e.shouldTakeBreak(requestCount) {
			e.breakPause()
		}
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("product_list page %d: %v", productPage, err))
			break
		}
		result.Products = append(result.Products, products...)
		logger.Info("[Compass] Got %d products on page %d", len(products), productPage)

		// Stop after first page or if no more results (like a real user browsing)
		if len(products) < 20 || productPage >= 5 {
			break
		}
		// Simulate: user scrolls down and loads next page
		e.humanDelay(3000, 8000)
		productPage++
	}

	// Step 3: Streamer rankings
	streamers, err := e.FetchStreamerRank(ctx, sessionID, startDate, endDate)
	requestCount++
	if e.shouldTakeBreak(requestCount) {
		e.breakPause()
	}
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("streamer_rank: %v", err))
	} else {
		result.Streamers = streamers
		logger.Info("[Compass] Got %d streamer rankings", len(streamers))
	}

	// Step 4: Funnel analysis
	funnel, err := e.FetchFunnelAnalysis(ctx, sessionID, startDate, endDate)
	requestCount++
	if e.shouldTakeBreak(requestCount) {
		e.breakPause()
	}
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("funnel: %v", err))
	} else {
		result.Funnel = funnel
		logger.Info("[Compass] Got %d funnel data points", len(funnel))
	}

	// Step 5: Orders (browse like a real user - a few pages)
	orderPage := 1
	for {
		orders, err := e.FetchOrderList(ctx, sessionID, startDate, endDate, orderPage, 50)
		requestCount++
		if e.shouldTakeBreak(requestCount) {
			e.breakPause()
		}
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("order_list page %d: %v", orderPage, err))
			break
		}
		result.Orders = append(result.Orders, orders...)
		logger.Info("[Compass] Got %d orders on page %d", len(orders), orderPage)

		if len(orders) < 50 || orderPage >= 3 {
			break
		}
		e.humanDelay(4000, 10000)
		orderPage++
	}

	result.TotalRequests = requestCount
	result.Duration = time.Since(start).Round(time.Second).String()

	logger.Info("[Compass] === Full sync complete: %d requests, %s, %d errors ===",
		requestCount, result.Duration, len(result.Errors))

	// Update last active timestamp
	now := time.Now()
	database.Get().Exec(context.Background(),
		`UPDATE compass_sessions SET last_active_at=$1 WHERE id=$2`, now, sessionID)

	return result, nil
}

// ─── Rate Limiting ───

func (e *CompassEngine) checkRateLimit(sessionID int64) bool {
	var dailyReqs, maxDaily int
	var nextAvail *time.Time
	err := database.Get().QueryRow(context.Background(),
		`SELECT daily_requests, max_daily_reqs, next_available_at FROM compass_sessions WHERE id=$1`,
		sessionID).Scan(&dailyReqs, &maxDaily, &nextAvail)
	if err != nil {
		return true
	}

	// Check if we need to wait until next_available_at (after a ban/cooldown)
	if nextAvail != nil && time.Now().Before(*nextAvail) {
		logger.Warn("[Compass] Session %d in cooldown until %v", sessionID, nextAvail)
		return false
	}

	if maxDaily > 0 && dailyReqs >= maxDaily {
		logger.Warn("[Compass] Session %d hit daily limit: %d/%d", sessionID, dailyReqs, maxDaily)
		return false
	}

	return true
}

func (e *CompassEngine) incrementRequestCount(sessionID int64) {
	database.Get().Exec(context.Background(),
		`UPDATE compass_sessions SET daily_requests = daily_requests + 1, last_active_at = NOW() WHERE id = $1`,
		sessionID)
}

// ─── Error Handling ───

func (e *CompassEngine) handleFetchError(sessionID int64, err error) error {
	errMsg := err.Error()

	// If session appears expired, mark it
	if strings.Contains(errMsg, "expired") || strings.Contains(errMsg, "302") ||
		strings.Contains(errMsg, "login") {
		e.markSessionStatus(sessionID, "expired")
	}

	// If banned, set cooldown
	if strings.Contains(errMsg, "403") || strings.Contains(errMsg, "banned") {
		e.markSessionCooldown(sessionID, 24*time.Hour)
	}

	// Increment fail counter
	database.Get().Exec(context.Background(),
		`UPDATE compass_sessions SET fail_count = fail_count + 1 WHERE id = $1`, sessionID)

	return fmt.Errorf("compass fetch error (session=%d): %w", sessionID, err)
}

func (e *CompassEngine) markSessionStatus(sessionID int64, status string) {
	database.Get().Exec(context.Background(),
		`UPDATE compass_sessions SET status=$1 WHERE id=$2`, status, sessionID)
	e.InvalidateSession(sessionID)
	logger.Warn("[Compass] Session %d marked as %s", sessionID, status)
}

func (e *CompassEngine) markSessionCooldown(sessionID int64, duration time.Duration) {
	nextAvail := time.Now().Add(duration)
	database.Get().Exec(context.Background(),
		`UPDATE compass_sessions SET status='cooldown', next_available_at=$1 WHERE id=$2`,
		nextAvail, sessionID)
	e.InvalidateSession(sessionID)
	logger.Warn("[Compass] Session %d in cooldown for %v", sessionID, duration)
}

// ─── Session Health Check ───

// CheckSessionHealth verifies a session cookie is still valid
func (e *CompassEngine) CheckSessionHealth(ctx context.Context, sessionID int64) (string, error) {
	sess, err := e.getSession(sessionID)
	if err != nil {
		return "error", err
	}

	// Make a lightweight request to check session validity
	body, err := e.doCompassRequest(ctx, sess, "/api/user/info", nil)
	if err != nil {
		return "expired", err
	}

	var resp struct {
		Code int `json:"code"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return "error", err
	}

	if resp.Code == 0 {
		// Reset fail count on successful check
		database.Get().Exec(context.Background(),
			`UPDATE compass_sessions SET status='active', fail_count=0 WHERE id=$1`, sessionID)
		return "active", nil
	}

	return "unknown", nil
}

// ─── Daily Reset ───
// Should be called once per day (e.g., at midnight) to reset daily counters

func (e *CompassEngine) ResetDailyCounters() {
	result, err := database.Get().Exec(context.Background(),
		`UPDATE compass_sessions SET daily_requests=0, fail_count=0 WHERE status='active'`)
	if err != nil {
		logger.Error("[Compass] Failed to reset daily counters: %v", err)
		return
	}
	rows, _ := result.RowsAffected()
	logger.Info("[Compass] Reset daily request counters for %d sessions", rows)
}

// suppress unused
var _ = logger.Info
