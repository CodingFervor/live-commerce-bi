package model

import "time"

// ═══ Douyin Compass (抖音电商罗盘) Data Models ═══

// CompassSession holds browser cookie session for compass access
type CompassSession struct {
	ID             int64      `json:"id" db:"id"`
	Name           string     `json:"name" db:"name"`
	Cookie         string     `json:"-" db:"cookie"`           // encrypted, never exposed via API
	ShopID         string     `json:"shop_id" db:"shop_id"`
	ShopName       string     `json:"shop_name" db:"shop_name"`
	UserAgent      string     `json:"user_agent" db:"user_agent"`
	ProxyURL       string     `json:"proxy_url,omitempty" db:"proxy_url"`
	Status         string     `json:"status" db:"status"`     // active, expired, banned, paused
	LastActiveAt   *time.Time `json:"last_active_at" db:"last_active_at"`
	DailyRequests  int        `json:"daily_requests" db:"daily_requests"`
	MaxDailyReqs   int        `json:"max_daily_reqs" db:"max_daily_reqs"`
	FailCount      int        `json:"fail_count" db:"fail_count"`
	NextAvailableAt *time.Time `json:"next_available_at" db:"next_available_at"`
	CreatedBy      int64      `json:"created_by" db:"created_by"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

type CompassSessionCreate struct {
	Name       string `json:"name" binding:"required"`
	Cookie     string `json:"cookie" binding:"required"`
	ShopID     string `json:"shop_id" binding:"required"`
	ShopName   string `json:"shop_name"`
	UserAgent  string `json:"user_agent"`
	ProxyURL   string `json:"proxy_url"`
	MaxDailyReqs int  `json:"max_daily_reqs"`
}

type CompassSessionUpdate struct {
	Name         *string `json:"name"`
	Cookie       *string `json:"cookie"`
	UserAgent    *string `json:"user_agent"`
	ProxyURL     *string `json:"proxy_url"`
	Status       *string `json:"status"`
	MaxDailyReqs *int    `json:"max_daily_reqs"`
}

// CompassTask represents a scheduled data collection task
type CompassTask struct {
	ID           int64      `json:"id" db:"id"`
	SessionID    int64      `json:"session_id" db:"session_id"`
	TaskType     string     `json:"task_type" db:"task_type"`     // live_overview, live_detail, product_list, product_detail, order_list, streamer_rank
	Params       string     `json:"params" db:"params"`           // JSON params
	Status       string     `json:"status" db:"status"`           // pending, running, completed, failed
	Result       string     `json:"result,omitempty" db:"result"` // JSON result summary
	RecordsCount int        `json:"records_count" db:"records_count"`
	ErrorMsg     string     `json:"error_msg,omitempty" db:"error_msg"`
	StartedAt    *time.Time `json:"started_at" db:"started_at"`
	CompletedAt  *time.Time `json:"completed_at" db:"completed_at"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
}

type CompassTaskCreate struct {
	SessionID int64  `json:"session_id" binding:"required"`
	TaskType  string `json:"task_type" binding:"required,oneof=live_overview live_detail product_list product_detail order_list streamer_rank funnel_analysis"`
	Params    string `json:"params"`
}

// ─── Compass Data Models (parsed from compass responses) ───

// CompassLiveOverview - 直播总览数据
type CompassLiveOverview struct {
	Date          string  `json:"date"`
	LiveCount     int     `json:"live_count"`
	TotalDuration int     `json:"total_duration"` // minutes
	TotalGMV      float64 `json:"total_gmv"`
	TotalOrders   int     `json:"total_orders"`
	TotalViews    int64   `json:"total_views"`
	TotalLikes    int64   `json:"total_likes"`
	TotalComments int64   `json:"total_comments"`
	AvgViewers    int     `json:"avg_viewers"`
	PeakViewers   int     `json:"peak_viewers"`
	NewFollowers  int     `json:"new_followers"`
	ConversionRate float64 `json:"conversion_rate"`
	AvgOrderValue  float64 `json:"avg_order_value"`
}

// CompassLiveDetail - 单场直播详情
type CompassLiveDetail struct {
	RoomID         string  `json:"room_id"`
	Title          string  `json:"title"`
	StreamerName   string  `json:"streamer_name"`
	Status         string  `json:"status"`
	StartTime      string  `json:"start_time"`
	EndTime        string  `json:"end_time"`
	Duration       int     `json:"duration"` // minutes
	TotalViews     int64   `json:"total_views"`
	PeakViewers    int     `json:"peak_viewers"`
	AvgViewers     int     `json:"avg_viewers"`
	TotalLikes     int64   `json:"total_likes"`
	TotalComments  int64   `json:"total_comments"`
	TotalShares    int64   `json:"total_shares"`
	NewFollowers   int     `json:"new_followers"`
	GMV            float64 `json:"gmv"`
	OrderCount     int     `json:"order_count"`
	ConversionRate float64 `json:"conversion_rate"`
	ProductCount   int     `json:"product_count"` // products shown
	TopProducts    []CompassProductBrief `json:"top_products"`
}

// CompassProductBrief - 商品简要信息
type CompassProductBrief struct {
	ProductID   string  `json:"product_id"`
	Title       string  `json:"title"`
	Category    string  `json:"category"`
	Price       float64 `json:"price"`
	GMV         float64 `json:"gmv"`
	OrderCount  int     `json:"order_count"`
	ViewCount   int64   `json:"view_count"`
	ClickCount  int64   `json:"click_count"`
	ConversionRate float64 `json:"conversion_rate"`
}

// CompassProductDetail - 商品详细数据
type CompassProductDetail struct {
	ProductID      string  `json:"product_id"`
	Title          string  `json:"title"`
	Category       string  `json:"category"`
	Price          float64 `json:"price"`
	Stock          int     `json:"stock"`
	TotalSales     int     `json:"total_sales"`
	TotalGMV       float64 `json:"total_gmv"`
	TotalViews     int64   `json:"total_views"`
	TotalClicks    int64   `json:"total_clicks"`
	TotalCart      int64   `json:"total_cart_adds"`
	ConversionRate float64 `json:"conversion_rate"`
	CartConvRate   float64 `json:"cart_conversion_rate"`
	RefundRate     float64 `json:"refund_rate"`
	AvgRating      float64 `json:"avg_rating"`
	ReviewCount    int     `json:"review_count"`
	Last7DaysSales []DailySale `json:"last_7_days_sales"`
}

type DailySale struct {
	Date  string `json:"date"`
	Sales int    `json:"sales"`
	GMV   float64 `json:"gmv"`
}

// CompassOrderItem - 订单数据
type CompassOrderItem struct {
	OrderID        string    `json:"order_id"`
	ProductID      string    `json:"product_id"`
	ProductTitle   string    `json:"product_title"`
	SKU            string    `json:"sku"`
	BuyerID        string    `json:"buyer_id"`
	Status         string    `json:"status"`
	Amount         float64   `json:"amount"`
	ActualAmount   float64   `json:"actual_amount"`
	Quantity       int       `json:"quantity"`
	Commission     float64   `json:"commission"`
	RefundAmount   float64   `json:"refund_amount"`
	LiveRoomID     string    `json:"live_room_id"`
	OrderedAt      time.Time `json:"ordered_at"`
	PaidAt         *time.Time `json:"paid_at"`
	ShippedAt      *time.Time `json:"shipped_at"`
	CompletedAt    *time.Time `json:"completed_at"`
}

// CompassStreamerRank - 主播排行
type CompassStreamerRank struct {
	Rank            int     `json:"rank"`
	StreamerID      string  `json:"streamer_id"`
	StreamerName    string  `json:"streamer_name"`
	Avatar          string  `json:"avatar"`
	LiveCount       int     `json:"live_count"`
	TotalGMV        float64 `json:"total_gmv"`
	TotalOrders     int     `json:"total_orders"`
	TotalViews      int64   `json:"total_views"`
	AvgViewers      int     `json:"avg_viewers"`
	AvgConversion   float64 `json:"avg_conversion"`
	AvgDuration     int     `json:"avg_duration"` // minutes
	FanGrowth       int     `json:"fan_growth"`
}

// CompassFunnelData - 漏斗分析数据
type CompassFunnelData struct {
	Date          string  `json:"date"`
	Exposure      int64   `json:"exposure"`       // 曝光
	Click         int64   `json:"click"`           // 点击
	CartAdd       int64   `json:"cart_add"`        // 加购
	Order         int64   `json:"order"`           // 下单
	Pay           int64   `json:"pay"`             // 支付
	ClickRate     float64 `json:"click_rate"`
	CartRate      float64 `json:"cart_rate"`
	OrderRate     float64 `json:"order_rate"`
	PayRate       float64 `json:"pay_rate"`
	OverallConvRate float64 `json:"overall_conv_rate"`
}

// CompassSyncStatus - 采集任务状态摘要
type CompassSyncStatus struct {
	SessionID       int64      `json:"session_id"`
	ShopName        string     `json:"shop_name"`
	Status          string     `json:"status"`
	LastSyncAt      *time.Time `json:"last_sync_at"`
	DailyRequests   int        `json:"daily_requests"`
	MaxDailyReqs    int        `json:"max_daily_reqs"`
	RecentTasks     []CompassTask `json:"recent_tasks"`
	NextAvailableAt *time.Time `json:"next_available_at"`
}
