package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
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
	"github.com/CodingFervor/live-commerce-bi/pkg/response"
)

func main() {
	// ─── Configuration ───
	cfgPath := "config/config.json"
	if envPath := os.Getenv("CONFIG_PATH"); envPath != "" {
		cfgPath = envPath
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// ─── Logger ───
	logger.Init(cfg.Log.Level)
	logger.Info("Starting Live Commerce BI Server v2.1...")

	// ─── JWT ───
	if cfg.JWT.Secret == "" {
		log.Fatal("JWT secret is required (set jwt.secret in config or JWT_SECRET env var)")
	}
	jwt.SetSecret(cfg.JWT.Secret)

	// ─── Database ───
	if err := database.Init(&cfg.Database); err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}
	defer database.Close()

	// ─── Redis ───
	if err := cache.Init(&cfg.Redis); err != nil {
		log.Fatalf("Failed to init redis: %v", err)
	}
	defer cache.Close()

	// ─── Gin Mode ───
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// ─── Background Services ───
	ctx, bgCancel := context.WithCancel(context.Background())
	defer bgCancel()

	// WebSocket Hub
	hub := service.NewHub()
	go hub.Run()
	logger.Info("WebSocket Hub started")

	// Redis Pub/Sub subscriber
	go service.RedisSubscriber(ctx)

	// Metrics aggregation scheduler (every 5 minutes)
	go func() {
		agg := service.NewMetricsAggregator()
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				aggCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
				now := time.Now()
				_ = agg.AggregateHourly(aggCtx, "", now.Add(-time.Hour))
				_ = agg.AggregateDaily(aggCtx, "", now.Add(-24*time.Hour))
				_ = agg.WarmupCache(aggCtx)
				cancel()
			}
		}
	}()

	// Intelligent Alert Engine
	if cfg.Alert.Enabled {
		alertEngine := service.NewAlertEngine()
		interval := time.Duration(cfg.Alert.EvaluateInterval) * time.Second
		go alertEngine.StartAlertScheduler(ctx, interval)
		logger.Info("Alert engine started (interval: %ds)", cfg.Alert.EvaluateInterval)
	}

	// Kafka event pipeline (optional)
	if cfg.Kafka.Enabled && len(cfg.Kafka.Brokers) > 0 {
		kafkaSvc := service.NewKafkaEventService(cfg.Kafka.Brokers, cfg.Kafka.GroupID)
		kafkaSvc.StartConsumers(ctx)
		defer kafkaSvc.Close()
		logger.Info("Kafka event pipeline started (brokers: %v)", cfg.Kafka.Brokers)
	}

	// Schedule dispatcher
	scheduler := service.NewScheduleDispatcher()
	go scheduler.Start(ctx)

	// Cache warmup on startup
	go func() {
		time.Sleep(5 * time.Second) // wait for services to stabilize
		qc := service.NewQueryCache()
		if err := qc.Warmup(ctx); err != nil {
			logger.Warn("Cache warmup failed: %v", err)
		}
	}()

	// ─── Router Setup ───
	r := gin.New()
	r.Use(response.RecoveryHandler())
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS())
	r.Use(middleware.RateLimit(200, time.Minute))

	// ─── System Endpoints ───
	r.GET("/health", healthCheck(hub))

	// Admin-only system info endpoints
	adminPublic := r.Group("/api/v1/system")
	adminPublic.Use(middleware.AuthRequired(), middleware.AdminRequired())
	{
		adminPublic.GET("/metrics", systemMetrics(hub))
		adminPublic.GET("/info", systemInfo())
	}

	// ─── API v1 ───
	v1 := r.Group("/api/v1")

	// Public: Auth
	authHandler := handler.NewAuthHandler()
	{
		v1.POST("/auth/login", authHandler.Login)
		v1.POST("/auth/register", authHandler.Register)
	}

	// Protected routes
	protected := v1.Group("")
	protected.Use(middleware.AuthRequired())
	protected.Use(middleware.DataPermission())
	protected.Use(middleware.AuditLogger())
	{
		protected.GET("/auth/profile", authHandler.GetProfile)

		// ─── Dashboard ───
		dashHandler := handler.NewDashboardHandler()
		{
			protected.GET("/dashboard/overview", dashHandler.Overview)
			protected.GET("/dashboard/realtime", dashHandler.Realtime)
			protected.GET("/dashboard/trend", dashHandler.Trend)
		}

		// Custom Dashboards
		{
			protected.POST("/dashboards", dashHandler.CreateDashboard)
			protected.GET("/dashboards", dashHandler.ListDashboards)
			protected.GET("/dashboards/:id", dashHandler.GetDashboard)
			protected.PUT("/dashboards/:id", dashHandler.UpdateDashboard)
			protected.DELETE("/dashboards/:id", dashHandler.DeleteDashboard)
			protected.POST("/dashboards/:id/widgets", dashHandler.CreateWidget)
			protected.GET("/dashboards/:id/widgets", dashHandler.ListWidgets)
			protected.PUT("/dashboards/:id/widgets/:wid", dashHandler.UpdateWidget)
			protected.DELETE("/dashboards/:id/widgets/:wid", dashHandler.DeleteWidget)
			protected.POST("/dashboards/:id/share", dashHandler.CreateDashboard)
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
		{
			protected.POST("/live-rooms", lrHandler.Create)
			protected.GET("/live-rooms", lrHandler.List)
			protected.GET("/live-rooms/:id", lrHandler.Get)
			protected.GET("/live-rooms/:id/metrics", lrHandler.GetMetrics)
			protected.GET("/live-rooms/:id/products", lrHandler.GetProducts)
			protected.GET("/live-rooms/:id/funnel", lrHandler.GetFunnel)
		}

		// ─── Streamers ───
		sHandler := handler.NewStreamerHandler()
		{
			protected.POST("/streamers", sHandler.Create)
			protected.GET("/streamers", sHandler.List)
			protected.GET("/streamers/:id", sHandler.Get)
			protected.PUT("/streamers/:id", sHandler.Update)
			protected.GET("/streamers/:id/performance", sHandler.Performance)
			protected.GET("/streamers/rankings", sHandler.Rankings)
		}

		// ─── Products ───
		pHandler := handler.NewProductHandler()
		{
			protected.POST("/products", pHandler.Create)
			protected.GET("/products", pHandler.List)
			protected.GET("/products/:id", pHandler.Get)
			protected.PUT("/products/:id", pHandler.Update)
			protected.GET("/products/rankings", pHandler.Rankings)
		}

		// ─── Orders ───
		oHandler := handler.NewOrderHandler()
		{
			protected.GET("/orders", oHandler.List)
			protected.GET("/orders/:id", oHandler.Get)
			protected.GET("/orders/stats", oHandler.Stats)
			protected.GET("/orders/revenue", oHandler.Revenue)
		}

		// ─── Analytics ───
		aHandler := handler.NewAnalyticsHandler()
		{
			protected.GET("/analytics/gmv", aHandler.GMV)
			protected.GET("/analytics/conversion", aHandler.Conversion)
			protected.GET("/analytics/platform-comparison", aHandler.PlatformComparison)
			protected.GET("/analytics/category-analysis", aHandler.CategoryAnalysis)
			protected.GET("/analytics/time-analysis", aHandler.TimeAnalysis)
			protected.GET("/analytics/funnel-analysis", aHandler.FunnelAnalysis)
			protected.GET("/viewers/realtime", aHandler.ViewerRealtime)
			protected.GET("/viewers/demographics", aHandler.ViewerDemographics)
			protected.GET("/viewers/engagement", aHandler.ViewerEngagement)
			protected.GET("/viewers/retention", aHandler.ViewerRetention)
		}

		// ─── Reports ───
		rHandler := handler.NewReportHandler()
		{
			protected.POST("/reports", rHandler.Create)
			protected.GET("/reports", rHandler.List)
			protected.GET("/reports/:id", rHandler.Get)
			protected.PUT("/reports/:id", rHandler.Update)
			protected.DELETE("/reports/:id", rHandler.Delete)
			protected.POST("/reports/:id/generate", rHandler.Generate)
			protected.GET("/reports/:id/download", rHandler.Download)
			protected.GET("/report-templates", rHandler.ListTemplates)
			protected.POST("/report-templates", rHandler.CreateTemplate)
		}

		// ─── Alerts ───
		alHandler := handler.NewAlertHandler()
		aeHandler := handler.NewAlertEngineHandler()
		{
			protected.POST("/alerts", alHandler.CreateRule)
			protected.GET("/alerts", alHandler.ListRules)
			protected.GET("/alerts/:id", alHandler.GetRule)
			protected.PUT("/alerts/:id", alHandler.UpdateRule)
			protected.DELETE("/alerts/:id", alHandler.DeleteRule)
			protected.GET("/alerts/:id/history", alHandler.GetHistory)
			protected.POST("/alerts/:id/test", alHandler.TestAlert)
			// Alert evaluation restricted to admins
			alertEval := protected.Group("/alerts")
			alertEval.Use(middleware.AdminRequired())
			{
				alertEval.POST("/evaluate", aeHandler.EvaluateAll)
				alertEval.POST("/:id/evaluate", aeHandler.EvaluateRule)
			}
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

		// ─── Audit Logs (Admin) ───
		auditHandler := handler.NewAuditHandler()
		auditLogs := protected.Group("/audit-logs")
		auditLogs.Use(middleware.AdminRequired())
		{
			auditLogs.GET("", auditHandler.ListAuditLogs)
			auditLogs.GET("/stats", auditHandler.GetAuditStats)
		}

		// ─── Data Export ───
		exportHandler := handler.NewExportHandler()
		{
			protected.POST("/exports", exportHandler.CreateExport)
			protected.GET("/exports", exportHandler.ListExports)
			protected.GET("/exports/:id", exportHandler.GetExportStatus)
			protected.GET("/exports/:id/download", exportHandler.DownloadExport)
			protected.POST("/exports/:id/cancel", exportHandler.CancelExport)
		}

		// ─── Advanced Analytics ───
		advHandler := handler.NewAdvancedAnalyticsHandler()
		{
			protected.GET("/advanced-analytics/cohort", advHandler.CohortAnalysis)
			protected.GET("/advanced-analytics/rfm", advHandler.RFMAnalysis)
			protected.GET("/advanced-analytics/forecast", advHandler.SalesForecast)
			protected.GET("/advanced-analytics/anomaly", advHandler.AnomalyDetection)
			protected.GET("/advanced-analytics/user-path", advHandler.UserPathAnalysis)
			protected.GET("/advanced-analytics/engagement-heatmap", advHandler.EngagementHeatmap)
			protected.POST("/advanced-analytics/olap", advHandler.OLAPQuery)
			protected.GET("/advanced-analytics/drill-down", advHandler.DrillDown)
			protected.GET("/advanced-analytics/period-comparison", advHandler.PeriodComparison)
			protected.GET("/advanced-analytics/target-comparison", advHandler.TargetComparison)
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

		// ─── Event Tracking ───
		evtHandler := handler.NewEventHandler()
		{
			v1.POST("/events/track", evtHandler.Track)
			v1.POST("/events/batch", evtHandler.BatchTrack)
			protected.GET("/events/funnel", evtHandler.GetFunnelEvents)
			protected.GET("/events/user-paths", evtHandler.GetUserPaths)
		}

		// ─── Aggregated Metrics ───
		metricsHandler := handler.NewMetricsAggHandler()
		{
			protected.GET("/metrics/hourly", metricsHandler.GetHourlyMetrics)
			protected.GET("/metrics/daily", metricsHandler.GetDailyMetrics)
			protected.GET("/metrics/streamer/:id/daily", metricsHandler.GetStreamerDaily)
		}

		// ─── Data Screen (Big Screen Display) ───
		screenHandler := handler.NewDataScreenHandler()
		{
			protected.GET("/screen/overview", screenHandler.Overview)
			protected.GET("/screen/realtime", screenHandler.Realtime)
			protected.GET("/screen/rankings", screenHandler.Rankings)
			protected.GET("/screen/geographic", screenHandler.Geographic)
		}

		// ─── AI Intelligence ───
		aiConfigHandler := handler.NewAIConfigHandler()
		aiChatHandler := handler.NewAIChatHandler()
		aiConfigAdmin := protected.Group("/ai/configs")
		aiConfigAdmin.Use(middleware.AdminRequired())
		{
			aiConfigAdmin.GET("", aiConfigHandler.ListAIConfigs)
			aiConfigAdmin.GET("/:id", aiConfigHandler.GetAIConfig)
			aiConfigAdmin.POST("", aiConfigHandler.CreateAIConfig)
			aiConfigAdmin.PUT("/:id", aiConfigHandler.UpdateAIConfig)
			aiConfigAdmin.DELETE("/:id", aiConfigHandler.DeleteAIConfig)
			aiConfigAdmin.POST("/:id/test", aiConfigHandler.TestAIConfig)
			aiConfigAdmin.POST("/:id/default", aiConfigHandler.SetDefaultAIConfig)
		}
		{
			protected.POST("/ai/chat", aiChatHandler.Chat)
			protected.POST("/ai/insights", aiChatHandler.GenerateInsights)
			protected.GET("/ai/conversations", aiChatHandler.ListConversations)
			protected.GET("/ai/conversations/:id", aiChatHandler.GetConversation)
			protected.DELETE("/ai/conversations/:id", aiChatHandler.DeleteConversation)
		}

		// ─── System Settings (Admin) ───
		settingHandler := handler.NewSystemSettingHandler()
		settingsAdmin := protected.Group("/settings")
		settingsAdmin.Use(middleware.AdminRequired())
		{
			settingsAdmin.GET("", settingHandler.ListSettings)
			settingsAdmin.GET("/:category/:key", settingHandler.GetSetting)
			settingsAdmin.PUT("/:category/:key", settingHandler.UpdateSetting)
			settingsAdmin.POST("/batch", settingHandler.BatchUpdateSettings)
			settingsAdmin.GET("/smtp", settingHandler.GetSMTPConfig)
			settingsAdmin.POST("/smtp", settingHandler.UpdateSMTPConfig)
			settingsAdmin.GET("/storage", settingHandler.GetStorageConfig)
			settingsAdmin.POST("/storage", settingHandler.UpdateStorageConfig)
			settingsAdmin.GET("/security", settingHandler.GetSecurityConfig)
			settingsAdmin.POST("/security", settingHandler.UpdateSecurityConfig)
			settingsAdmin.GET("/system-info", settingHandler.GetSystemInfo)
		}
	}

	// ─── WebSocket (requires auth) ───
	wsHandler := handler.NewWebSocketHandler()
	wsGroup := r.Group("/ws")
	wsGroup.Use(middleware.AuthRequired())
	{
		wsGroup.GET("", wsHandler.ServeWS)
		wsGroup.GET("/stats", wsHandler.WSStats)
		wsGroup.GET("/room/:room_id", wsHandler.SubscribeRoom)
	}

	// ─── Start Server ───
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  120 * time.Second,
		MaxHeaderBytes: cfg.Server.MaxHeaderBytes,
	}

	go func() {
		logger.Info("Server listening on %s (mode=%s)", addr, cfg.Server.Mode)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// ─── Graceful Shutdown ───
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	// Stop background services
	bgCancel()
	scheduler.Stop()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced shutdown: %v", err)
	}
	logger.Info("Server exited gracefully")
}

// ─── Endpoint Handlers ───

func healthCheck(hub *service.Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
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
			"status":     status,
			"service":    "live-commerce-bi",
			"version":    "2.1.0",
			"db":         dbStatus,
			"redis":      redisStatus,
			"ws_clients": hub.ClientCount(),
			"timestamp":  time.Now().Unix(),
		})
	}
}

func systemMetrics(hub *service.Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"ws_clients":  hub.ClientCount(),
			"ws_rooms":    hub.RoomCount(),
			"goroutines":  runtime.NumGoroutine(),
		})
	}
}

func systemInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"name":       "Live Commerce BI",
			"version":    "2.1.0",
			"go_version": runtime.Version(),
			"platform":   runtime.GOOS + "/" + runtime.GOARCH,
			"num_cpu":    runtime.NumCPU(),
		})
	}
}
