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

	// Create router
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS())
	r.Use(middleware.RateLimit(200, time.Minute))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"service":   "live-commerce-bi",
			"timestamp": time.Now().Unix(),
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
	}

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
