package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/CodingFervor/live-commerce-bi/internal/cache"
	"github.com/CodingFervor/live-commerce-bi/pkg/logger"
)

// ═══ Query Cache Layer ═══
// Transparent caching for expensive analytical queries
// Reference: Alibaba DataV caching strategy

type QueryCache struct {
	defaultTTL time.Duration
}

func NewQueryCache() *QueryCache {
	return &QueryCache{
		defaultTTL: 5 * time.Minute,
	}
}

// Get retrieves cached data, returns nil if not found or expired
func (qc *QueryCache) Get(ctx context.Context, key string, result interface{}) bool {
	rdb := cache.Get()
	if rdb == nil {
		return false
	}

	data, err := cache.GetJSON(ctx, qc.key(key))
	if err != nil {
		return false
	}

	if err := json.Unmarshal([]byte(data), result); err != nil {
		return false
	}
	return true
}

// Set stores data in cache with configurable TTL
func (qc *QueryCache) Set(ctx context.Context, key string, data interface{}, ttl ...time.Duration) {
	rdb := cache.Get()
	if rdb == nil {
		return
	}

	cacheTTL := qc.defaultTTL
	if len(ttl) > 0 {
		cacheTTL = ttl[0]
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return
	}

	if err := cache.SetJSON(ctx, qc.key(key), string(jsonData), cacheTTL); err != nil {
		logger.Error("Cache set failed for %s: %v", key, err)
	}
}

// Invalidate removes a cached entry
func (qc *QueryCache) Invalidate(ctx context.Context, key string) {
	rdb := cache.Get()
	if rdb == nil {
		return
	}
	cache.Del(ctx, qc.key(key))
}

// InvalidatePattern removes all cached entries matching a prefix
func (qc *QueryCache) InvalidatePattern(ctx context.Context, prefix string) {
	rdb := cache.Get()
	if rdb == nil {
		return
	}
	// Scan and delete keys matching the pattern
	iter := rdb.Scan(ctx, 0, qc.key(prefix)+"*", 100).Iterator()
	for iter.Next(ctx) {
		rdb.Del(ctx, iter.Val())
	}
}

// Warmup preloads frequently accessed data into cache
func (qc *QueryCache) Warmup(ctx context.Context) error {
	logger.Info("Starting cache warmup...")

	// Warmup 1: Platform summary
	qc.warmPlatformSummary(ctx)

	// Warmup 2: Today's KPIs
	qc.warmTodayKPIs(ctx)

	// Warmup 3: Top streamers
	qc.warmTopStreamers(ctx)

	// Warmup 4: Active live rooms
	qc.warmActiveRooms(ctx)

	logger.Info("Cache warmup completed")
	return nil
}

func (qc *QueryCache) warmPlatformSummary(ctx context.Context) {
	// Handled by MetricsAggregator.WarmupCache
	agg := NewMetricsAggregator()
	agg.WarmupCache(ctx)
}

func (qc *QueryCache) warmTodayKPIs(ctx context.Context) {
	// Pre-compute today's KPIs and cache them
	screen := NewDataScreenService()
	data, err := screen.ScreenOverview(ctx)
	if err == nil {
		qc.Set(ctx, "screen:overview", data, 5*time.Minute)
	}
}

func (qc *QueryCache) warmTopStreamers(ctx context.Context) {
	screen := NewDataScreenService()
	data, err := screen.ScreenRankings(ctx, "streamer_gmv", "today", 10)
	if err == nil {
		qc.Set(ctx, "rankings:streamer_gmv:today", data, 5*time.Minute)
	}
}

func (qc *QueryCache) warmActiveRooms(ctx context.Context) {
	screen := NewDataScreenService()
	data, err := screen.ScreenRealtime(ctx)
	if err == nil {
		qc.Set(ctx, "screen:realtime", data, 10*time.Second)
	}
}

func (qc *QueryCache) key(k string) string {
	return "query_cache:" + k
}

// Cache TTL presets for different data types
var (
	CacheTTLRealtime = 10 * time.Second  // live metrics
	CacheTTLMinute   = 1 * time.Minute   // near-realtime dashboards
	CacheTTLDefault  = 5 * time.Minute   // standard analytics
	CacheTTLHour     = 1 * time.Hour     // hourly aggregations
	CacheTTLDaily    = 24 * time.Hour    // daily reports
)
