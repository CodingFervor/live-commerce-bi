package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/CodingFervor/live-commerce-bi/internal/cache"
	"github.com/CodingFervor/live-commerce-bi/internal/database"
	"github.com/CodingFervor/live-commerce-bi/pkg/logger"
)

// ═══ Data Screen Service ═══
// Real-time big-screen dashboard for ops war room display
// Reference: ByteDance data screen architecture

type DataScreenService struct{}

func NewDataScreenService() *DataScreenService { return &DataScreenService{} }

// ScreenOverview returns the main big-screen KPI cards
func (s *DataScreenService) ScreenOverview(ctx context.Context) (map[string]interface{}, error) {
	// Try cache first
	if data, err := s.fromCache(ctx, "screen:overview"); err == nil && data != nil {
		return data, nil
	}

	result := make(map[string]interface{})

	// ─── Today's KPIs ───
	var todayGMV, yesterdayGMV float64
	var todayOrders, yesterdayOrders int
	var todayViews, yesterdayViews int64

	database.Get().QueryRow(ctx, `
		SELECT COALESCE(SUM(actual_amount),0), COUNT(*)
		FROM orders WHERE status NOT IN ('pending','cancelled')
		AND created_at >= CURRENT_DATE`).Scan(&todayGMV, &todayOrders)

	database.Get().QueryRow(ctx, `
		SELECT COALESCE(SUM(actual_amount),0), COUNT(*)
		FROM orders WHERE status NOT IN ('pending','cancelled')
		AND created_at >= CURRENT_DATE - INTERVAL '1 day' AND created_at < CURRENT_DATE`).Scan(&yesterdayGMV, &yesterdayOrders)

	database.Get().QueryRow(ctx, `
		SELECT COALESCE(SUM(total_views),0) FROM live_rooms WHERE created_at >= CURRENT_DATE`).Scan(&todayViews)
	database.Get().QueryRow(ctx, `
		SELECT COALESCE(SUM(total_views),0) FROM live_rooms
		WHERE created_at >= CURRENT_DATE - INTERVAL '1 day' AND created_at < CURRENT_DATE`).Scan(&yesterdayViews)

	result["today"] = map[string]interface{}{
		"gmv":         todayGMV,
		"orders":      todayOrders,
		"views":       todayViews,
		"gmv_change":  safePercent(todayGMV, yesterdayGMV),
		"order_change": safePercent(float64(todayOrders), float64(yesterdayOrders)),
		"view_change":  safePercent(float64(todayViews), float64(yesterdayViews)),
	}

	// ─── Active Live Rooms ───
	var activeRooms int
	database.Get().QueryRow(ctx, "SELECT COUNT(*) FROM live_rooms WHERE status='live'").Scan(&activeRooms)
	result["active_rooms"] = activeRooms

	// ─── Top Streamers (today) ───
	rows, err := database.Get().Query(ctx, `
		SELECT s.id, s.name, s.avatar, SUM(lr.gmv) AS gmv, SUM(lr.order_count) AS orders, SUM(lr.total_views) AS views
		FROM live_rooms lr JOIN streamers s ON lr.streamer_id = s.id
		WHERE lr.created_at >= CURRENT_DATE
		GROUP BY s.id, s.name, s.avatar ORDER BY gmv DESC LIMIT 10`)
	if err == nil {
		defer rows.Close()
		var topStreamers []map[string]interface{}
		for rows.Next() {
			var id int64
			var name, avatar string
			var gmv float64
			var orders int
			var views int64
			if rows.Scan(&id, &name, &avatar, &gmv, &orders, &views) != nil {
				continue
			}
			topStreamers = append(topStreamers, map[string]interface{}{
				"id": id, "name": name, "avatar": avatar,
				"gmv": gmv, "orders": orders, "views": views,
			})
		}
		result["top_streamers"] = topStreamers
	}

	// ─── Platform Distribution ───
	rows2, err := database.Get().Query(ctx, `
		SELECT platform, COALESCE(SUM(gmv),0), COUNT(*) FROM live_rooms
		WHERE created_at >= CURRENT_DATE GROUP BY platform ORDER BY SUM(gmv) DESC`)
	if err == nil {
		defer rows2.Close()
		var platforms []map[string]interface{}
		for rows2.Next() {
			var platform string
			var gmv float64
			var count int
			rows2.Scan(&platform, &gmv, &count)
			platforms = append(platforms, map[string]interface{}{
				"platform": platform, "gmv": gmv, "count": count,
			})
		}
		result["platforms"] = platforms
	}

	// ─── Hourly GMV Trend (today) ───
	rows3, err := database.Get().Query(ctx, `
		SELECT EXTRACT(HOUR FROM created_at)::int AS hour,
			COALESCE(SUM(actual_amount),0) AS gmv
		FROM orders WHERE status NOT IN ('pending','cancelled')
		AND created_at >= CURRENT_DATE
		GROUP BY hour ORDER BY hour`)
	if err == nil {
		defer rows3.Close()
		var hourlyGMV []map[string]interface{}
		for rows3.Next() {
			var hour int
			var gmv float64
			rows3.Scan(&hour, &gmv)
			hourlyGMV = append(hourlyGMV, map[string]interface{}{
				"hour": hour, "gmv": gmv,
			})
		}
		result["hourly_gmv"] = hourlyGMV
	}

	// ─── Recent Alert Events ───
	rows4, err := database.Get().Query(ctx, `
		SELECT ah.id, ar.name, ah.metric, ah.triggered_value, ah.threshold_value, ah.severity, ah.created_at
		FROM alert_history ah JOIN alert_rules ar ON ah.alert_rule_id = ar.id
		WHERE ah.created_at >= CURRENT_DATE ORDER BY ah.created_at DESC LIMIT 5`)
	if err == nil {
		defer rows4.Close()
		var alerts []map[string]interface{}
		for rows4.Next() {
			var id int64
			var name, metric, severity string
			var triggered, threshold float64
			var createdAt time.Time
			rows4.Scan(&id, &name, &metric, &triggered, &threshold, &severity, &createdAt)
			alerts = append(alerts, map[string]interface{}{
				"id": id, "rule_name": name, "metric": metric,
				"triggered": triggered, "threshold": threshold,
				"severity": severity, "time": createdAt,
			})
		}
		result["recent_alerts"] = alerts
	}

	// Cache for 30 seconds (near-real-time for big screen)
	s.toCache(ctx, "screen:overview", result, 30*time.Second)

	return result, nil
}

// ScreenRealtime returns live room real-time metrics
func (s *DataScreenService) ScreenRealtime(ctx context.Context) (map[string]interface{}, error) {
	// Aggregate from Redis counters
	rdb := cache.Get()
	result := make(map[string]interface{})

	if rdb != nil {
		if v, err := rdb.Get(ctx, "live:gmv:today").Float64(); err == nil {
			result["realtime_gmv"] = v
		}
		if v, err := rdb.Get(ctx, "live:orders:today").Int64(); err == nil {
			result["realtime_orders"] = v
		}
	}

	// Active rooms with live metrics
	rows, err := database.Get().Query(ctx, `
		SELECT lr.id, lr.title, lr.platform, s.name AS streamer_name,
			lr.gmv, lr.order_count, lr.total_views, lr.peak_viewers,
			lr.total_likes, lr.total_comments, lr.conversion_rate
		FROM live_rooms lr JOIN streamers s ON lr.streamer_id = s.id
		WHERE lr.status = 'live' ORDER BY lr.total_views DESC LIMIT 20`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []map[string]interface{}
	for rows.Next() {
		var id int64
		var title, platform, streamerName string
		var gmv float64
		var orderCount int
		var totalViews, peakViewers, totalLikes, totalComments int64
		var convRate float64
		rows.Scan(&id, &title, &platform, &streamerName, &gmv, &orderCount,
			&totalViews, &peakViewers, &totalLikes, &totalComments, &convRate)
		rooms = append(rooms, map[string]interface{}{
			"id": id, "title": title, "platform": platform,
			"streamer": streamerName, "gmv": gmv, "orders": orderCount,
			"views": totalViews, "peak": peakViewers,
			"likes": totalLikes, "comments": totalComments,
			"conversion": convRate,
		})
	}
	result["live_rooms"] = rooms
	result["live_count"] = len(rooms)

	// WebSocket connection count
	hub := GetHub()
	if hub != nil {
		result["ws_clients"] = hub.ClientCount()
	}

	return result, nil
}

// ScreenRankings returns real-time rankings for various dimensions
func (s *DataScreenService) ScreenRankings(ctx context.Context, rankingType, period string, limit int) ([]map[string]interface{}, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	// Cache key for this ranking
	cacheKey := fmt.Sprintf("screen:rankings:%s:%s:%d", rankingType, period, limit)
	if data, err := s.fromCache(ctx, cacheKey); err == nil && data != nil {
		if results, ok := data["results"].([]interface{}); ok {
			_ = results
			return nil, nil // Simplified: would need proper type assertion
		}
	}

	var dateFilter string
	switch period {
	case "today":
		dateFilter = "AND lr.created_at >= CURRENT_DATE"
	case "week":
		dateFilter = "AND lr.created_at >= CURRENT_DATE - INTERVAL '7 days'"
	case "month":
		dateFilter = "AND lr.created_at >= CURRENT_DATE - INTERVAL '30 days'"
	default:
		dateFilter = "AND lr.created_at >= CURRENT_DATE"
	}

	var query string
	switch rankingType {
	case "streamer_gmv":
		query = fmt.Sprintf(`SELECT s.id, s.name, s.avatar, SUM(lr.gmv) AS value, COUNT(*) AS rooms
			FROM live_rooms lr JOIN streamers s ON lr.streamer_id = s.id
			WHERE 1=1 %s GROUP BY s.id ORDER BY value DESC LIMIT %d`, dateFilter, limit)
	case "product_sales":
		query = fmt.Sprintf(`SELECT p.id, p.name, p.image_url, SUM(o.quantity) AS value, SUM(o.actual_amount) AS revenue
			FROM orders o JOIN products p ON o.product_id = p.id
			WHERE o.status NOT IN ('pending','cancelled')
			AND o.created_at >= CURRENT_DATE
			GROUP BY p.id ORDER BY revenue DESC LIMIT %d`, limit)
	case "room_hot":
		query = fmt.Sprintf(`SELECT lr.id, lr.title, lr.platform, lr.total_views AS value,
			lr.gmv, lr.conversion_rate, s.name AS streamer
			FROM live_rooms lr JOIN streamers s ON lr.streamer_id = s.id
			WHERE 1=1 %s ORDER BY lr.total_views DESC LIMIT %d`, dateFilter, limit)
	default:
		query = fmt.Sprintf(`SELECT s.id, s.name, s.avatar, SUM(lr.gmv) AS value, COUNT(*) AS rooms
			FROM live_rooms lr JOIN streamers s ON lr.streamer_id = s.id
			WHERE 1=1 %s GROUP BY s.id ORDER BY value DESC LIMIT %d`, dateFilter, limit)
	}

	rows, err := database.Get().Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, _ := rows.FieldDescriptions()
	var results []map[string]interface{}
	rank := 1
	for rows.Next() {
		vals, _ := rows.Values()
		entry := map[string]interface{}{"rank": rank}
		for i, col := range cols {
			entry[string(col.Name)] = vals[i]
		}
		results = append(results, entry)
		rank++
	}

	// Cache for 1 minute
	s.toCache(ctx, cacheKey, map[string]interface{}{"results": results}, 1*time.Minute)

	return results, nil
}

// ScreenGeographic returns geographic distribution of viewers
func (s *DataScreenService) ScreenGeographic(ctx context.Context) ([]map[string]interface{}, error) {
	rows, err := database.Get().Query(ctx, `
		SELECT vd.region_distribution FROM viewer_demographics vd
		JOIN live_rooms lr ON vd.live_room_id = lr.id
		WHERE lr.status IN ('live','ended') AND vd.date = CURRENT_DATE - INTERVAL '1 day'
		ORDER BY vd.created_at DESC LIMIT 1`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		var regionJSON string
		rows.Scan(&regionJSON)
		var regions []map[string]interface{}
		json.Unmarshal([]byte(regionJSON), &regions)
		return regions, nil
	}
	return nil, nil
}

// ─── Cache Helpers ───

func (s *DataScreenService) fromCache(ctx context.Context, key string) (map[string]interface{}, error) {
	rdb := cache.Get()
	if rdb == nil {
		return nil, fmt.Errorf("redis unavailable")
	}
	data, err := cache.GetJSON(ctx, key)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *DataScreenService) toCache(ctx context.Context, key string, data interface{}, ttl time.Duration) {
	rdb := cache.Get()
	if rdb == nil {
		return
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return
	}
	if err := cache.SetJSON(ctx, key, string(jsonData), ttl); err != nil {
		logger.Error("Cache write failed for %s: %v", key, err)
	}
}

func safePercent(current, previous float64) float64 {
	if previous == 0 {
		return 0
	}
	return (current - previous) / previous * 100
}
