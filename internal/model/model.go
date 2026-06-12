package model

import "time"

// ─── User ───
type User struct {
	ID         int64     `json:"id" db:"id"`
	Username   string    `json:"username" db:"username"`
	Email      string    `json:"email" db:"email"`
	Password   string    `json:"-" db:"password"`
	Role       string    `json:"role" db:"role"`
	Avatar     string    `json:"avatar" db:"avatar"`
	Status     string    `json:"status" db:"status"`
	LastLoginAt *time.Time `json:"last_login_at" db:"last_login_at"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role"`
}

type LoginResponse struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
	User      User   `json:"user"`
}

// ─── DataSource ───
type DataSource struct {
	ID           int64      `json:"id" db:"id"`
	Name         string     `json:"name" db:"name"`
	Platform     string     `json:"platform" db:"platform"`
	Config       string     `json:"config" db:"config"`
	Status       string     `json:"status" db:"status"`
	LastSyncAt   *time.Time `json:"last_sync_at" db:"last_sync_at"`
	SyncInterval int        `json:"sync_interval" db:"sync_interval"`
	CreatedBy    int64      `json:"created_by" db:"created_by"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

type DataSourceCreate struct {
	Name         string `json:"name" binding:"required"`
	Platform     string `json:"platform" binding:"required,oneof=douyin kuaishou taobao_live jd_live pdd_live"`
	Config       string `json:"config"`
	SyncInterval int    `json:"sync_interval"`
}

type DataSourceUpdate struct {
	Name         *string `json:"name"`
	Config       *string `json:"config"`
	SyncInterval *int    `json:"sync_interval"`
	Status       *string `json:"status"`
}

// ─── Streamer ───
type Streamer struct {
	ID            int64     `json:"id" db:"id"`
	Name          string    `json:"name" db:"name"`
	Platform      string    `json:"platform" db:"platform"`
	PlatformID    string    `json:"platform_id" db:"platform_id"`
	Avatar        string    `json:"avatar" db:"avatar"`
	FollowerCount int64     `json:"follower_count" db:"follower_count"`
	Category      string    `json:"category" db:"category"`
	Tags          []string  `json:"tags" db:"tags"`
	Status        string    `json:"status" db:"status"`
	DataSourceID  int64     `json:"data_source_id" db:"data_source_id"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

type StreamerCreate struct {
	Name         string   `json:"name" binding:"required"`
	Platform     string   `json:"platform" binding:"required"`
	PlatformID   string   `json:"platform_id"`
	Avatar       string   `json:"avatar"`
	Category     string   `json:"category"`
	Tags         []string `json:"tags"`
	DataSourceID int64    `json:"data_source_id"`
}

type StreamerPerformance struct {
	Streamer       Streamer `json:"streamer"`
	TotalLiveRooms int      `json:"total_live_rooms"`
	TotalGMV       float64  `json:"total_gmv"`
	TotalOrders    int      `json:"total_orders"`
	TotalViews     int64    `json:"total_views"`
	AvgConversion  float64  `json:"avg_conversion"`
	AvgViewers     int      `json:"avg_viewers"`
	TotalDuration  int      `json:"total_duration"` // seconds
}

type StreamerRanking struct {
	Rank         int     `json:"rank"`
	StreamerID   int64   `json:"streamer_id"`
	StreamerName string  `json:"streamer_name"`
	Platform     string  `json:"platform"`
	Avatar       string  `json:"avatar"`
	Category     string  `json:"category"`
	GMV          float64 `json:"gmv"`
	OrderCount   int     `json:"order_count"`
	ViewerCount  int64   `json:"viewer_count"`
	ConversionRate float64 `json:"conversion_rate"`
}

// ─── LiveRoom ───
type LiveRoom struct {
	ID             int64      `json:"id" db:"id"`
	StreamerID     int64      `json:"streamer_id" db:"streamer_id"`
	Platform       string     `json:"platform" db:"platform"`
	RoomID         string     `json:"room_id" db:"room_id"`
	Title          string     `json:"title" db:"title"`
	Status         string     `json:"status" db:"status"`
	StartedAt      *time.Time `json:"started_at" db:"started_at"`
	EndedAt        *time.Time `json:"ended_at" db:"ended_at"`
	Duration       int        `json:"duration" db:"duration"`
	PeakViewers    int        `json:"peak_viewers" db:"peak_viewers"`
	AvgViewers     int        `json:"avg_viewers" db:"avg_viewers"`
	TotalViews     int64      `json:"total_views" db:"total_views"`
	TotalLikes     int64      `json:"total_likes" db:"total_likes"`
	TotalComments  int64      `json:"total_comments" db:"total_comments"`
	TotalShares    int64      `json:"total_shares" db:"total_shares"`
	GMV            float64    `json:"gmv" db:"gmv"`
	OrderCount     int        `json:"order_count" db:"order_count"`
	ProductCount   int        `json:"product_count" db:"product_count"`
	ConversionRate float64    `json:"conversion_rate" db:"conversion_rate"`
	Tags           []string   `json:"tags" db:"tags"`
	DataSourceID   int64      `json:"data_source_id" db:"data_source_id"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
	// Joined fields
	StreamerName string `json:"streamer_name,omitempty" db:"streamer_name"`
}

type LiveRoomCreate struct {
	StreamerID   int64    `json:"streamer_id" binding:"required"`
	Platform     string   `json:"platform" binding:"required"`
	RoomID       string   `json:"room_id"`
	Title        string   `json:"title" binding:"required"`
	Tags         []string `json:"tags"`
	DataSourceID int64    `json:"data_source_id"`
}

type LiveRoomMetrics struct {
	RoomID         int64   `json:"room_id"`
	ConcurrentIndex int    `json:"concurrent_viewers"`
	PeakViewers    int     `json:"peak_viewers"`
	TotalViews     int64   `json:"total_views"`
	TotalLikes     int64   `json:"total_likes"`
	TotalComments  int64   `json:"total_comments"`
	GMV            float64 `json:"gmv"`
	OrderCount     int     `json:"order_count"`
	ConversionRate float64 `json:"conversion_rate"`
	AvgWatchTime   int     `json:"avg_watch_time"` // seconds
}

// ─── Product ───
type Product struct {
	ID                int64     `json:"id" db:"id"`
	Name              string    `json:"name" db:"name"`
	Platform          string    `json:"platform" db:"platform"`
	PlatformProductID string    `json:"platform_product_id" db:"platform_product_id"`
	Category          string    `json:"category" db:"category"`
	Brand             string    `json:"brand" db:"brand"`
	Price             float64   `json:"price" db:"price"`
	OriginalPrice     float64   `json:"original_price" db:"original_price"`
	LivePrice         float64   `json:"live_price" db:"live_price"`
	ImageURL          string    `json:"image_url" db:"image_url"`
	Description       string    `json:"description" db:"description"`
	Tags              []string  `json:"tags" db:"tags"`
	Status            string    `json:"status" db:"status"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

type ProductCreate struct {
	Name              string   `json:"name" binding:"required"`
	Platform          string   `json:"platform" binding:"required"`
	PlatformProductID string   `json:"platform_product_id"`
	Category          string   `json:"category"`
	Brand             string   `json:"brand"`
	Price             float64  `json:"price"`
	OriginalPrice     float64  `json:"original_price"`
	LivePrice         float64  `json:"live_price"`
	ImageURL          string   `json:"image_url"`
	Description       string   `json:"description"`
	Tags              []string `json:"tags"`
}

type ProductRanking struct {
	Rank         int     `json:"rank"`
	ProductID    int64   `json:"product_id"`
	ProductName  string  `json:"product_name"`
	Brand        string  `json:"brand"`
	Category     string  `json:"category"`
	ImageURL     string  `json:"image_url"`
	Price        float64 `json:"price"`
	TotalSold    int     `json:"total_sold"`
	TotalRevenue float64 `json:"total_revenue"`
	Appearances  int     `json:"appearances"` // in how many live rooms
	ConversionRate float64 `json:"conversion_rate"`
}

type ProductAnalytics struct {
	Product       Product    `json:"product"`
	DailySales    []DailySale `json:"daily_sales"`
	PriceHistory  []PricePoint `json:"price_history"`
	TopLiveRooms  []LiveRoom  `json:"top_live_rooms"`
}

type DailySale struct {
	Date    string  `json:"date"`
	Orders  int     `json:"orders"`
	Revenue float64 `json:"revenue"`
	Quantity int    `json:"quantity"`
}

type PricePoint struct {
	Date  string  `json:"date"`
	Price float64 `json:"price"`
}

// ─── Order ───
type Order struct {
	ID              int64      `json:"id" db:"id"`
	OrderNo         string     `json:"order_no" db:"order_no"`
	Platform        string     `json:"platform" db:"platform"`
	PlatformOrderID string     `json:"platform_order_id" db:"platform_order_id"`
	LiveRoomID      int64      `json:"live_room_id" db:"live_room_id"`
	ProductID       int64      `json:"product_id" db:"product_id"`
	StreamerID      int64      `json:"streamer_id" db:"streamer_id"`
	BuyerID         string     `json:"buyer_id" db:"buyer_id"`
	Quantity        int        `json:"quantity" db:"quantity"`
	UnitPrice       float64    `json:"unit_price" db:"unit_price"`
	TotalAmount     float64    `json:"total_amount" db:"total_amount"`
	DiscountAmount  float64    `json:"discount_amount" db:"discount_amount"`
	ActualAmount    float64    `json:"actual_amount" db:"actual_amount"`
	CommissionRate  float64    `json:"commission_rate" db:"commission_rate"`
	Commission      float64    `json:"commission" db:"commission"`
	Status          string     `json:"status" db:"status"`
	PaidAt          *time.Time `json:"paid_at" db:"paid_at"`
	ShippedAt       *time.Time `json:"shipped_at" db:"shipped_at"`
	CompletedAt     *time.Time `json:"completed_at" db:"completed_at"`
	RefundedAt      *time.Time `json:"refunded_at" db:"refunded_at"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
	// Joined fields
	ProductName  string `json:"product_name,omitempty" db:"product_name"`
	StreamerName string `json:"streamer_name,omitempty" db:"streamer_name"`
	RoomTitle    string `json:"room_title,omitempty" db:"room_title"`
}

type OrderStats struct {
	TotalOrders    int     `json:"total_orders"`
	TotalGMV       float64 `json:"total_gmv"`
	TotalActual    float64 `json:"total_actual"`
	TotalCommission float64 `json:"total_commission"`
	TotalRefund    float64 `json:"total_refund"`
	AvgOrderValue  float64 `json:"avg_order_value"`
	CompletedRate  float64 `json:"completed_rate"`
	RefundRate     float64 `json:"refund_rate"`
}

type RevenueRecord struct {
	ID           int64      `json:"id" db:"id"`
	Date         string     `json:"date" db:"date"`
	Platform     string     `json:"platform" db:"platform"`
	LiveRoomID   int64      `json:"live_room_id" db:"live_room_id"`
	StreamerID   int64      `json:"streamer_id" db:"streamer_id"`
	GMV          float64    `json:"gmv" db:"gmv"`
	ActualRevenue float64   `json:"actual_revenue" db:"actual_revenue"`
	Commission   float64    `json:"commission" db:"commission"`
	RefundAmount float64    `json:"refund_amount" db:"refund_amount"`
	OrderCount   int        `json:"order_count" db:"order_count"`
	AvgOrderValue float64   `json:"avg_order_value" db:"avg_order_value"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

// ─── Viewer Metrics ───
type ViewerMetrics struct {
	ID               int64     `json:"id" db:"id"`
	LiveRoomID       int64     `json:"live_room_id" db:"live_room_id"`
	Timestamp        time.Time `json:"timestamp" db:"timestamp"`
	ConcurrentIndex  int       `json:"concurrent_viewers" db:"concurrent_viewers"`
	NewFollowers     int       `json:"new_followers" db:"new_followers"`
	Likes            int       `json:"likes" db:"likes"`
	Comments         int       `json:"comments" db:"comments"`
	Shares           int       `json:"shares" db:"shares"`
	GiftsCount       int       `json:"gifts_count" db:"gifts_count"`
	GiftsValue       float64   `json:"gifts_value" db:"gifts_value"`
	AvgWatchTime     int       `json:"avg_watch_time" db:"avg_watch_time"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
}

type ViewerDemographics struct {
	ID                 int64     `json:"id" db:"id"`
	LiveRoomID         int64     `json:"live_room_id" db:"live_room_id"`
	Date               string    `json:"date" db:"date"`
	AgeDistribution    string    `json:"age_distribution" db:"age_distribution"`
	GenderDistribution string    `json:"gender_distribution" db:"gender_distribution"`
	RegionDistribution string    `json:"region_distribution" db:"region_distribution"`
	DeviceDistribution string    `json:"device_distribution" db:"device_distribution"`
	TopCities          string    `json:"top_cities" db:"top_cities"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
}

type ViewerEngagement struct {
	TotalViews      int64   `json:"total_views"`
	UniqueViewers   int64   `json:"unique_viewers"`
	AvgWatchTime    int     `json:"avg_watch_time"` // seconds
	EngagementRate  float64 `json:"engagement_rate"`
	LikeRate        float64 `json:"like_rate"`
	CommentRate     float64 `json:"comment_rate"`
	ShareRate       float64 `json:"share_rate"`
	FollowRate      float64 `json:"follow_rate"`
	GiftRate        float64 `json:"gift_rate"`
	AvgGiftValue    float64 `json:"avg_gift_value"`
}

type ViewerRetention struct {
	Cohort     string  `json:"cohort"`
	Day0       float64 `json:"day_0"`
	Day1       float64 `json:"day_1"`
	Day3       float64 `json:"day_3"`
	Day7       float64 `json:"day_7"`
	Day14      float64 `json:"day_14"`
	Day30      float64 `json:"day_30"`
}

// ─── Conversion Funnel ───
type ConversionFunnel struct {
	ID          int64     `json:"id" db:"id"`
	LiveRoomID  int64     `json:"live_room_id" db:"live_room_id"`
	Timestamp   time.Time `json:"timestamp" db:"timestamp"`
	Impressions int64     `json:"impressions" db:"impressions"`
	Clicks      int64     `json:"clicks" db:"clicks"`
	AddToCarts  int64     `json:"add_to_carts" db:"add_to_carts"`
	Orders      int64     `json:"orders" db:"orders"`
	Payments    int64     `json:"payments" db:"payments"`
	ClickRate   float64   `json:"click_rate" db:"click_rate"`
	CartRate    float64   `json:"cart_rate" db:"cart_rate"`
	OrderRate   float64   `json:"order_rate" db:"order_rate"`
	PaymentRate float64   `json:"payment_rate" db:"payment_rate"`
	OverallRate float64   `json:"overall_rate" db:"overall_rate"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// ─── Dashboard ───
type Dashboard struct {
	ID          int64     `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Type        string    `json:"type" db:"type"`
	Layout      string    `json:"layout" db:"layout"`
	IsDefault   bool      `json:"is_default" db:"is_default"`
	OwnerID     int64     `json:"owner_id" db:"owner_id"`
	Status      string    `json:"status" db:"status"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type DashboardWidget struct {
	ID              int64     `json:"id" db:"id"`
	DashboardID     int64     `json:"dashboard_id" db:"dashboard_id"`
	Title           string    `json:"title" db:"title"`
	Type            string    `json:"type" db:"type"`
	DataSource      string    `json:"data_source" db:"data_source"`
	Config          string    `json:"config" db:"config"`
	Position        string    `json:"position" db:"position"`
	RefreshInterval int       `json:"refresh_interval" db:"refresh_interval"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

type DashboardCreate struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Layout      string `json:"layout"`
	IsDefault   bool   `json:"is_default"`
}

type WidgetCreate struct {
	Title           string `json:"title" binding:"required"`
	Type            string `json:"type" binding:"required"`
	DataSource      string `json:"data_source"`
	Config          string `json:"config"`
	Position        string `json:"position"`
	RefreshInterval int    `json:"refresh_interval"`
}

// ─── Report ───
type ReportTemplate struct {
	ID          int64     `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Type        string    `json:"type" db:"type"`
	Sections    string    `json:"sections" db:"sections"`
	Thumbnail   string    `json:"thumbnail" db:"thumbnail"`
	IsPublic    bool      `json:"is_public" db:"is_public"`
	CreatedBy   int64     `json:"created_by" db:"created_by"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type Report struct {
	ID            int64      `json:"id" db:"id"`
	Name          string     `json:"name" db:"name"`
	TemplateID    int64      `json:"template_id" db:"template_id"`
	Type          string     `json:"type" db:"type"`
	Status        string     `json:"status" db:"status"`
	Config        string     `json:"config" db:"config"`
	Schedule      string     `json:"schedule" db:"schedule"`
	DateRangeStart *string   `json:"date_range_start" db:"date_range_start"`
	DateRangeEnd   *string   `json:"date_range_end" db:"date_range_end"`
	Filters       string     `json:"filters" db:"filters"`
	OutputFormat  string     `json:"output_format" db:"output_format"`
	FilePath      string     `json:"file_path" db:"file_path"`
	GeneratedAt   *time.Time `json:"generated_at" db:"generated_at"`
	GeneratedBy   int64      `json:"generated_by" db:"generated_by"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

type ReportCreate struct {
	Name           string `json:"name" binding:"required"`
	TemplateID     int64  `json:"template_id"`
	Type           string `json:"type" binding:"required,oneof=daily weekly monthly custom one_time"`
	Config         string `json:"config"`
	Schedule       string `json:"schedule"`
	DateRangeStart string `json:"date_range_start"`
	DateRangeEnd   string `json:"date_range_end"`
	Filters        string `json:"filters"`
	OutputFormat   string `json:"output_format"`
}

// ─── Alert ───
type AlertRule struct {
	ID             int64     `json:"id" db:"id"`
	Name           string    `json:"name" db:"name"`
	Description    string    `json:"description" db:"description"`
	Metric         string    `json:"metric" db:"metric"`
	Condition      string    `json:"condition" db:"condition"`
	Threshold      float64   `json:"threshold" db:"threshold"`
	Timeframe      string    `json:"timeframe" db:"timeframe"`
	Severity       string    `json:"severity" db:"severity"`
	NotifyChannels string    `json:"notify_channels" db:"notify_channels"`
	NotifyConfig   string    `json:"notify_config" db:"notify_config"`
	Cooldown       int       `json:"cooldown" db:"cooldown"`
	IsEnabled      bool      `json:"is_enabled" db:"is_enabled"`
	CreatedBy      int64     `json:"created_by" db:"created_by"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

type AlertRuleCreate struct {
	Name           string  `json:"name" binding:"required"`
	Description    string  `json:"description"`
	Metric         string  `json:"metric" binding:"required,oneof=gmv viewers conversion_rate order_count refund_rate peak_viewers avg_watch_time engagement_rate"`
	Condition      string  `json:"condition" binding:"required,oneof=gt lt eq gte lte change_pct_gt change_pct_lt"`
	Threshold      float64 `json:"threshold" binding:"required"`
	Timeframe      string  `json:"timeframe"`
	Severity       string  `json:"severity"`
	NotifyChannels string  `json:"notify_channels"`
	NotifyConfig   string  `json:"notify_config"`
	Cooldown       int     `json:"cooldown"`
}

type AlertHistory struct {
	ID              int64      `json:"id" db:"id"`
	AlertRuleID     int64      `json:"alert_rule_id" db:"alert_rule_id"`
	Metric          string     `json:"metric" db:"metric"`
	TriggeredValue  float64    `json:"triggered_value" db:"triggered_value"`
	ThresholdValue  float64    `json:"threshold_value" db:"threshold_value"`
	Message         string     `json:"message" db:"message"`
	Severity        string     `json:"severity" db:"severity"`
	Notified        bool       `json:"notified" db:"notified"`
	Acknowledged    bool       `json:"acknowledged" db:"acknowledged"`
	AcknowledgedBy  *int64     `json:"acknowledged_by" db:"acknowledged_by"`
	AcknowledgedAt  *time.Time `json:"acknowledged_at" db:"acknowledged_at"`
	ResolvedAt      *time.Time `json:"resolved_at" db:"resolved_at"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	RuleName        string     `json:"rule_name,omitempty" db:"rule_name"`
}

// ─── Analytics Models ───
type OverviewStats struct {
	TotalGMV         float64          `json:"total_gmv"`
	TotalOrders      int              `json:"total_orders"`
	TotalLiveRooms   int              `json:"total_live_rooms"`
	TotalStreamers   int              `json:"total_streamers"`
	TotalViews       int64            `json:"total_views"`
	AvgConversion    float64          `json:"avg_conversion"`
	AvgOrderValue    float64          `json:"avg_order_value"`
	TotalCommission  float64          `json:"total_commission"`
	ActiveLiveRooms  int              `json:"active_live_rooms"`
	GMVChange        float64          `json:"gmv_change"` // percent change
	OrderChange      float64          `json:"order_change"`
	ViewerChange     float64          `json:"viewer_change"`
	TopPlatforms     []PlatformStats  `json:"top_platforms"`
	RecentLiveRooms  []LiveRoom       `json:"recent_live_rooms"`
}

type PlatformStats struct {
	Platform    string  `json:"platform"`
	GMV         float64 `json:"gmv"`
	Orders      int     `json:"orders"`
	Views       int64   `json:"views"`
	LiveRooms   int     `json:"live_rooms"`
	Conversion  float64 `json:"conversion_rate"`
	Share       float64 `json:"share"` // market share %
}

type GMVAnalytics struct {
	Period    string  `json:"period"`
	GMV       float64 `json:"gmv"`
	Orders    int     `json:"orders"`
	AvgValue  float64 `json:"avg_order_value"`
	Change    float64 `json:"change_pct"`
}

type ConversionAnalysis struct {
	Stage       string  `json:"stage"`
	Count       int64   `json:"count"`
	Rate        float64 `json:"rate"`
	DropOff     float64 `json:"drop_off_rate"`
}

type CategoryAnalysis struct {
	Category     string  `json:"category"`
	GMV          float64 `json:"gmv"`
	Orders       int     `json:"orders"`
	Products     int     `json:"products"`
	AvgPrice     float64 `json:"avg_price"`
	Conversion   float64 `json:"conversion_rate"`
	Growth       float64 `json:"growth_pct"`
}

type TimeAnalysis struct {
	Hour        int     `json:"hour"`
	GMV         float64 `json:"gmv"`
	Orders      int     `json:"orders"`
	Viewers     int64   `json:"viewers"`
	Engagement  float64 `json:"engagement"`
}

type PlatformComparison struct {
	Platform       string  `json:"platform"`
	GMV            float64 `json:"gmv"`
	Orders         int     `json:"orders"`
	Viewers        int64   `json:"viewers"`
	ConversionRate float64 `json:"conversion_rate"`
	AvgOrderValue  float64 `json:"avg_order_value"`
	StreamerCount  int     `json:"streamer_count"`
	LiveRoomCount  int     `json:"live_room_count"`
	TopCategory    string  `json:"top_category"`
}

type CohortData struct {
	CohortDate string    `json:"cohort_date"`
	Users      int       `json:"users"`
	Retention  []float64 `json:"retention"` // day 1..7
	Revenue    float64   `json:"revenue"`
}

// ─── Sync ───
type SyncLog struct {
	ID               int64      `json:"id" db:"id"`
	DataSourceID     int64      `json:"data_source_id" db:"data_source_id"`
	SyncType         string     `json:"sync_type" db:"sync_type"`
	Status           string     `json:"status" db:"status"`
	StartedAt        *time.Time `json:"started_at" db:"started_at"`
	CompletedAt      *time.Time `json:"completed_at" db:"completed_at"`
	RecordsProcessed int        `json:"records_processed" db:"records_processed"`
	RecordsFailed    int        `json:"records_failed" db:"records_failed"`
	ErrorMessage     string     `json:"error_message" db:"error_message"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
}
