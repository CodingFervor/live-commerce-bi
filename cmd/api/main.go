package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/CodingFervor/live-commerce-bi/internal/cache"
	"github.com/CodingFervor/live-commerce-bi/internal/config"
	"github.com/CodingFervor/live-commerce-bi/internal/database"
	"github.com/CodingFervor/live-commerce-bi/internal/handler"
	"github.com/CodingFervor/live-commerce-bi/internal/middleware"
	"github.com/CodingFervor/live-commerce-bi/internal/service"
	"github.com/CodingFervor/live-commerce-bi/pkg/jwt"
	"github.com/CodingFervor/live-commerce-bi/pkg/logger"
)

func main() {
	// Load config
	cfgPath := "config/config.json"
	if envPath := os.Getenv("CONFIG_PATH"); envPath != "" {
		cfgPath = envPath
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Init logger
	logger.Init(cfg.Log.Level)
	logger.Info("Starting Live Commerce BI Server...")

	// Init JWT
	jwt.SetSecret(cfg.JWT.Secret)

	// Init Database
	if err := database.Init(&cfg.Database); err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}
	defer database.Close()

	// Init Redis
	if err := cache.Init(&cfg.Redis); err != nil {
		log.Fatalf("Failed to init redis: %v", err)
	}
	defer cache.Close()

	// Set Gin mode
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Start WebSocket Hub
	hub := service.NewHub()
	go hub.Run()
	logger.Info("WebSocket Hub started")

	// Start Redis subscriber for real-time push
	go service.RedisSubscriber(context.Background())

	// Start metrics aggregation scheduler
	go func() {
		agg := service.NewMetricsAggregator()
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			now := time.Now()
			_ = agg.AggregateHourly(ctx, "", now.Add(-time.Hour))
			_ = agg.AggregateDaily(ctx, "", now.Add(-24*time.Hour))
			_ = agg.WarmupCache(ctx)
			cancel()
		}
	}()

	// Create router
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS())
	r.Use(middleware.RateLimit(200, time.Minute))

	// Health check with dependency status
	r.GET("/health", func(c *gin.Context) {
		status := "ok"
		dbStatus := "up"
		redisStatus := "up"

		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()

		if err := database.Get().Ping(ctx); err != nil {
			dbStatus = "down"
			status = "degraded"
		}
		if err := cache.Get().Ping(ctx).Err(); err != nil {
			redisStatus = "down"
			status = "degraded"
		}

		code := http.StatusOK
		if status == "degraded" {
			code = http.StatusServiceUnavailable
		}
		c.JSON(code, gin.H{
			"status":    status,
			"service":   "live-commerce-bi",
			"version":   "2.0.0",
			"db":        dbStatus,
			"redis":     redisStatus,
			"ws_clients": hub.ClientCount(),
			"timestamp": time.Now().Unix(),
		})
	})

	r.GET("/metrics", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"ws_clients":  hub.ClientCount(),
			"ws_rooms":    hub.RoomCount(),
			"goroutines":  "N/A",
		})
	})

	// API v1 group
	v1 := r.Group("/api/v1")

	// ─── Auth (public) ───
	authHandler := handler.NewAuthHandler()
	auth := v1.Group("/auth")
	{
		auth.POST("/login", authHandler.Login)
		auth.POST("/register", authHandler.Register)
	}

	// ─── Protected routes ───
	protected := v1.Group("")
	protected.Use(middleware.AuthRequired())
	protected.Use(middleware.AuditLogger())
	{
		// Auth profile
		protected.GET("/auth/profile", authHandler.GetProfile)

		// ─── Dashboard ───
		dashHandler := handler.NewDashboardHandler()
		dashboard := protected.Group("/dashboard")
		{
			dashboard.GET("/overview", dashHandler.Overview)
			dashboard.GET("/realtime", dashHandler.Realtime)
			dashboard.GET("/trend", dashHandler.Trend)
		}

		// Custom dashboards
		dashboards := protected.Group("/dashboards")
		{
			dashboards.POST("", dashHandler.CreateDashboard)
			dashboards.GET("", dashHandler.ListDashboards)
			dashboards.GET("/:id", dashHandler.GetDashboard)
			dashboards.PUT("/:id", dashHandler.UpdateDashboard)
			dashboards.DELETE("/:id", dashHandler.DeleteDashboard)
			dashboards.POST("/:id/widgets", dashHandler.CreateWidget)
			dashboards.GET("/:id/widgets", dashHandler.ListWidgets)
			dashboards.PUT("/:id/widgets/:wid", dashHandler.UpdateWidget)
			dashboards.DELETE("/:id/widgets/:wid", dashHandler.DeleteWidget)
		}

		// ─── Data Sources ───
		dsHandler := handler.NewDataSourceHandler()
		dataSources := protected.Group("/datasources")
		dataSources.Use(middleware.RoleRequired("admin", "analyst"))
		{
			dataSources.POST("", dsHandler.Create)
			dataSources.GET("", dsHandler.List)
			dataSources.GET("/:id", dsHandler.Get)
			dataSources.PUT("/:id", dsHandler.Update)
			dataSources.DELETE("/:id", dsHandler.Delete)
			dataSources.POST("/:id/sync", dsHandler.Sync)
			dataSources.GET("/:id/status", dsHandler.Status)
		}

		// ─── Live Rooms ───
		lrHandler := handler.NewLiveRoomHandler()
		liveRooms := protected.Group("/live-rooms")
		{
			liveRooms.POST("", lrHandler.Create)
			liveRooms.GET("", lrHandler.List)
			liveRooms.GET("/:id", lrHandler.Get)
			liveRooms.GET("/:id/metrics", lrHandler.GetMetrics)
			liveRooms.GET("/:id/products", lrHandler.GetProducts)
			liveRooms.GET("/:id/funnel", lrHandler.GetFunnel)
		}

		// ─── Streamers ───
		sHandler := handler.NewStreamerHandler()
		streamers := protected.Group("/streamers")
		{
			streamers.POST("", sHandler.Create)
			streamers.GET("", sHandler.List)
			streamers.GET("/:id", sHandler.Get)
			streamers.PUT("/:id", sHandler.Update)
			streamers.GET("/:id/performance", sHandler.Performance)
			streamers.GET("/rankings", sHandler.Rankings)
		}

		// ─── Products ───
		pHandler := handler.NewProductHandler()
		products := protected.Group("/products")
		{
			products.POST("", pHandler.Create)
			products.GET("", pHandler.List)
			products.GET("/:id", pHandler.Get)
			products.PUT("/:id", pHandler.Update)
			products.GET("/rankings", pHandler.Rankings)
		}

		// ─── Orders ───
		oHandler := handler.NewOrderHandler()
		orders := protected.Group("/orders")
		{
			orders.GET("", oHandler.List)
			orders.GET("/:id", oHandler.Get)
			orders.GET("/stats", oHandler.Stats)
			orders.GET("/revenue", oHandler.Revenue)
		}

		// ─── Analytics ───
		aHandler := handler.NewAnalyticsHandler()
		analytics := protected.Group("/analytics")
		{
			analytics.GET("/gmv", aHandler.GMV)
			analytics.GET("/conversion", aHandler.Conversion)
			analytics.GET("/platform-comparison", aHandler.PlatformComparison)
			analytics.GET("/category-analysis", aHandler.CategoryAnalysis)
			analytics.GET("/time-analysis", aHandler.TimeAnalysis)
			analytics.GET("/funnel-analysis", aHandler.FunnelAnalysis)
		}

		// ─── Viewers ───
		viewers := protected.Group("/viewers")
		{
			viewers.GET("/realtime", aHandler.ViewerRealtime)
			viewers.GET("/demographics", aHandler.ViewerDemographics)
			viewers.GET("/engagement", aHandler.ViewerEngagement)
			viewers.GET("/retention", aHandler.ViewerRetention)
		}

		// ─── Reports ───
		rHandler := handler.NewReportHandler()
		reports := protected.Group("/reports")
		{
			reports.POST("", rHandler.Create)
			reports.GET("", rHandler.List)
			reports.GET("/:id", rHandler.Get)
			reports.PUT("/:id", rHandler.Update)
			reports.DELETE("/:id", rHandler.Delete)
			reports.POST("/:id/generate", rHandler.Generate)
			reports.GET("/:id/download", rHandler.Download)
		}

		reportTemplates := protected.Group("/report-templates")
		{
			reportTemplates.GET("", rHandler.ListTemplates)
			reportTemplates.POST("", rHandler.CreateTemplate)
		}

		// ─── Alerts ───
		alHandler := handler.NewAlertHandler()
		alerts := protected.Group("/alerts")
		{
			alerts.POST("", alHandler.CreateRule)
			alerts.GET("", alHandler.ListRules)
			alerts.GET("/:id", alHandler.GetRule)
			alerts.PUT("/:id", alHandler.UpdateRule)
			alerts.DELETE("/:id", alHandler.DeleteRule)
			alerts.GET("/:id/history", alHandler.GetHistory)
			alerts.POST("/:id/test", alHandler.TestAlert)
		}

		// ─── Organizations (Admin) ───
		orgHandler := handler.NewOrganizationHandler()
		orgs := protected.Group("/organizations")
		orgs.Use(middleware.AdminRequired())
		{
			orgs.POST("", orgHandler.CreateOrg)
			orgs.GET("", orgHandler.ListOrgs)
			orgs.GET("/:id", orgHandler.GetOrg)
			orgs.PUT("/:id", orgHandler.UpdateOrg)
			orgs.POST("/:id/departments", orgHandler.CreateDept)
			orgs.GET("/:id/departments/tree", orgHandler.GetDeptTree)
			orgs.PUT("/:id/departments/:did", orgHandler.UpdateDept)
			orgs.DELETE("/:id/departments/:did", orgHandler.DeleteDept)
		}

		// ─── RBAC (Admin) ───
		rbacHandler := handler.NewRBACHandler()
		rbac := protected.Group("/rbac")
		rbac.Use(middleware.AdminRequired())
		{
			rbac.GET("/permissions", rbacHandler.ListPermissions)
			rbac.POST("/roles", rbacHandler.CreateRole)
			rbac.GET("/roles", rbacHandler.ListRoles)
			rbac.PUT("/roles/:id", rbacHandler.UpdateRole)
			rbac.POST("/roles/:id/users/:uid", rbacHandler.AssignRole)
			rbac.GET("/me/permissions", rbacHandler.GetCurrentUserPermissions)
		}

		// ─── Audit Logs ───
		auditHandler := handler.NewAuditHandler()
		auditLogs := protected.Group("/audit-logs")
		{
			auditLogs.GET("", auditHandler.ListAuditLogs)
			auditLogs.GET("/stats", auditHandler.GetAuditStats)
		}

		// ─── Data Export ───
		exportHandler := handler.NewExportHandler()
		exports := protected.Group("/exports")
		{
			exports.POST("", exportHandler.CreateExport)
			exports.GET("", exportHandler.ListExports)
			exports.GET("/:id", exportHandler.GetExportStatus)
			exports.GET("/:id/download", exportHandler.DownloadExport)
			exports.POST("/:id/cancel", exportHandler.CancelExport)
		}

		// ─── Advanced Analytics ───
		advHandler := handler.NewAdvancedAnalyticsHandler()
		advanced := protected.Group("/advanced-analytics")
		{
			advanced.GET("/cohort", advHandler.CohortAnalysis)
			advanced.GET("/rfm", advHandler.RFMAnalysis)
			advanced.GET("/forecast", advHandler.SalesForecast)
			advanced.GET("/anomaly", advHandler.AnomalyDetection)
			advanced.GET("/user-path", advHandler.UserPathAnalysis)
			advanced.GET("/engagement-heatmap", advHandler.EngagementHeatmap)
			advanced.POST("/olap", advHandler.OLAPQuery)
			advanced.GET("/drill-down", advHandler.DrillDown)
			advanced.GET("/period-comparison", advHandler.PeriodComparison)
			advanced.GET("/target-comparison", advHandler.TargetComparison)
		}

		// ─── Data Quality ───
		dqHandler := handler.NewDataQualityHandler()
		quality := protected.Group("/data-quality")
		quality.Use(middleware.RoleRequired("admin", "analyst"))
		{
			quality.POST("/rules", dqHandler.CreateRule)
			quality.GET("/rules", dqHandler.ListRules)
			quality.POST("/rules/:id/check", dqHandler.RunCheck)
			quality.GET("/rules/:id/results", dqHandler.GetResults)
		}

		// ─── Event Tracking (Public-facing, no auth for SDK) ───
		evtHandler := handler.NewEventHandler()
		{
			v1.POST("/events/track", evtHandler.Track)
			v1.POST("/events/batch", evtHandler.BatchTrack)
			protected.GET("/events/funnel", evtHandler.GetFunnelEvents)
			protected.GET("/events/user-paths", evtHandler.GetUserPaths)
		}

		// ─── Aggregated Metrics ───
		metricsHandler := handler.NewMetricsAggHandler()
		metricsAgg := protected.Group("/metrics")
		{
			metricsAgg.GET("/hourly", metricsHandler.GetHourlyMetrics)
			metricsAgg.GET("/daily", metricsHandler.GetDailyMetrics)
			metricsAgg.GET("/streamer/:id/daily", metricsHandler.GetStreamerDaily)
		}

		// ─── Dashboard Sharing ───
		shares := protected.Group("/dashboards")
		{
			shares.POST("/:id/share", dashHandler.CreateDashboard) // reuse existing
		}
	}

	// ─── WebSocket (public endpoint, auth via query param) ───
	wsHandler := handler.NewWebSocketHandler()
	r.GET("/ws", wsHandler.ServeWS)
	r.GET("/ws/stats", wsHandler.WSStats)
	r.GET("/ws/room/:room_id", wsHandler.SubscribeRoom)

	// ─── Start Server ───
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Graceful shutdown
	go func() {
		logger.Info("Server listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced shutdown: %v", err)
	}
	logger.Info("Server exited gracefully")
}
