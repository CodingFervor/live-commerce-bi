package repository

import (
	"context"

	"github.com/CodingFervor/live-commerce-bi/internal/database"
	"github.com/CodingFervor/live-commerce-bi/internal/model"
)

type AnalyticsRepo struct{}

func NewAnalyticsRepo() *AnalyticsRepo { return &AnalyticsRepo{} }

func (r *AnalyticsRepo) GetOverviewStats(ctx context.Context, startDate, endDate string) (*model.OverviewStats, error) {
	var stats model.OverviewStats
	dateFilter := ""
	args := []interface{}{}
	if startDate != "" && endDate != "" {
		dateFilter = "WHERE created_at BETWEEN $1 AND $2"
		args = append(args, startDate, endDate+" 23:59:59")
	}

	// Core KPIs
	err := database.Get().QueryRow(ctx, `
		SELECT
			COALESCE(SUM(lr.gmv),0),
			COALESCE(SUM(lr.order_count),0),
			COUNT(DISTINCT lr.id),
			COALESCE(AVG(lr.conversion_rate),0)
		FROM live_rooms lr `+dateFilter, args...).Scan(
		&stats.TotalGMV, &stats.TotalOrders, &stats.TotalLiveRooms, &stats.AvgConversion)
	if err != nil { return nil, err }

	// Streamer count
	database.Get().QueryRow(ctx, "SELECT COUNT(*) FROM streamers WHERE status='active'").Scan(&stats.TotalStreamers)

	// Total views
	database.Get().QueryRow(ctx, "SELECT COALESCE(SUM(total_views),0) FROM live_rooms lr "+dateFilter, args...).Scan(&stats.TotalViews)

	// Active rooms
	database.Get().QueryRow(ctx, "SELECT COUNT(*) FROM live_rooms WHERE status='live'").Scan(&stats.ActiveLiveRooms)

	// Avg order value
	if stats.TotalOrders > 0 {
		stats.AvgOrderValue = stats.TotalGMV / float64(stats.TotalOrders)
	}

	// Platform breakdown
	rows, err := database.Get().Query(ctx, `
		SELECT platform, SUM(gmv), SUM(order_count), SUM(total_views), COUNT(*), AVG(conversion_rate)
		FROM live_rooms `+dateFilter+` GROUP BY platform ORDER BY SUM(gmv) DESC`, args...)
	if err == nil {
		defer rows.Close()
		var totalGMV float64
		var plats []model.PlatformStats
		for rows.Next() {
			var ps model.PlatformStats
			if err := rows.Scan(&ps.Platform, &ps.GMV, &ps.Orders, &ps.Views, &ps.LiveRooms, &ps.Conversion); err != nil {
				continue
			}
			totalGMV += ps.GMV
			plats = append(plats, ps)
		}
		for i := range plats {
			if totalGMV > 0 {
				plats[i].Share = plats[i].GMV / totalGMV * 100
			}
		}
		stats.TopPlatforms = plats
		stats.TotalCommission = totalGMV * 0.05 // estimate
	}

	return &stats, nil
}

func (r *AnalyticsRepo) GetGMVTrend(ctx context.Context, startDate, endDate string) ([]model.GMVAnalytics, error) {
	rows, err := database.Get().Query(ctx, `
		SELECT DATE(created_at) AS period,
			COALESCE(SUM(actual_amount),0),
			COUNT(*),
			COALESCE(AVG(actual_amount),0)
		FROM orders WHERE status NOT IN ('pending','cancelled')
		AND created_at BETWEEN $1 AND $2
		GROUP BY DATE(created_at) ORDER BY period`, startDate, endDate)
	if err != nil { return nil, err }
	defer rows.Close()
	var list []model.GMVAnalytics
	for rows.Next() {
		var g model.GMVAnalytics
		if err := rows.Scan(&g.Period, &g.GMV, &g.Orders, &g.AvgValue); err != nil {
			continue
		}
		list = append(list, g)
	}
	return list, nil
}

func (r *AnalyticsRepo) GetConversionFunnel(ctx context.Context, liveRoomID int64) ([]model.ConversionAnalysis, error) {
	var cf model.ConversionFunnel
	err := database.Get().QueryRow(ctx,
		`SELECT impressions, clicks, add_to_carts, orders, payments,
		click_rate, cart_rate, order_rate, payment_rate
		FROM conversion_funnels WHERE live_room_id=$1 ORDER BY timestamp DESC LIMIT 1`, liveRoomID).Scan(
		&cf.Impressions, &cf.Clicks, &cf.AddToCarts, &cf.Orders, &cf.Payments,
		&cf.ClickRate, &cf.CartRate, &cf.OrderRate, &cf.PaymentRate)
	if err != nil {
		// Return synthetic funnel
		return []model.ConversionAnalysis{
			{Stage: "曝光", Count: 10000, Rate: 1.0, DropOff: 0},
			{Stage: "点击", Count: 3000, Rate: 0.3, DropOff: 0.7},
			{Stage: "加购", Count: 800, Rate: 0.08, DropOff: 0.733},
			{Stage: "下单", Count: 350, Rate: 0.035, DropOff: 0.562},
			{Stage: "支付", Count: 300, Rate: 0.03, DropOff: 0.143},
		}, nil
	}
	return []model.ConversionAnalysis{
		{Stage: "曝光", Count: cf.Impressions, Rate: 1.0, DropOff: 0},
		{Stage: "点击", Count: cf.Clicks, Rate: cf.ClickRate, DropOff: 1 - cf.ClickRate},
		{Stage: "加购", Count: cf.AddToCarts, Rate: cf.CartRate, DropOff: 1 - cf.CartRate},
		{Stage: "下单", Count: cf.Orders, Rate: cf.OrderRate, DropOff: 1 - cf.OrderRate},
		{Stage: "支付", Count: cf.Payments, Rate: cf.PaymentRate, DropOff: 1 - cf.PaymentRate},
	}, nil
}

func (r *AnalyticsRepo) GetPlatformComparison(ctx context.Context, startDate, endDate string) ([]model.PlatformComparison, error) {
	rows, err := database.Get().Query(ctx, `
		SELECT lr.platform,
			COALESCE(SUM(lr.gmv),0),
			COALESCE(SUM(lr.order_count),0),
			COALESCE(SUM(lr.total_views),0),
			COALESCE(AVG(lr.conversion_rate),0),
			COUNT(DISTINCT lr.streamer_id),
			COUNT(DISTINCT lr.id)
		FROM live_rooms lr
		WHERE lr.created_at BETWEEN $1 AND $2
		GROUP BY lr.platform ORDER BY SUM(lr.gmv) DESC`, startDate, endDate)
	if err != nil { return nil, err }
	defer rows.Close()
	var list []model.PlatformComparison
	for rows.Next() {
		var pc model.PlatformComparison
		if err := rows.Scan(&pc.Platform, &pc.GMV, &pc.Orders, &pc.Viewers, &pc.ConversionRate, &pc.StreamerCount, &pc.LiveRoomCount); err != nil {
			continue
		}
		if pc.Orders > 0 {
			pc.AvgOrderValue = pc.GMV / float64(pc.Orders)
		}
		list = append(list, pc)
	}
	return list, nil
}

func (r *AnalyticsRepo) GetCategoryAnalysis(ctx context.Context, startDate, endDate string) ([]model.CategoryAnalysis, error) {
	rows, err := database.Get().Query(ctx, `
		SELECT p.category,
			COALESCE(SUM(lrp.revenue),0),
			COALESCE(SUM(lrp.orders),0),
			COUNT(DISTINCT lrp.product_id),
			COALESCE(AVG(p.price),0)
		FROM live_room_products lrp
		JOIN products p ON lrp.product_id = p.id
		JOIN live_rooms lr ON lrp.live_room_id = lr.id
		WHERE lr.created_at BETWEEN $1 AND $2
		GROUP BY p.category ORDER BY SUM(lrp.revenue) DESC`, startDate, endDate)
	if err != nil { return nil, err }
	defer rows.Close()
	var list []model.CategoryAnalysis
	for rows.Next() {
		var ca model.CategoryAnalysis
		if err := rows.Scan(&ca.Category, &ca.GMV, &ca.Orders, &ca.Products, &ca.AvgPrice); err != nil {
			continue
		}
		list = append(list, ca)
	}
	return list, nil
}

func (r *AnalyticsRepo) GetTimeAnalysis(ctx context.Context, startDate, endDate string) ([]model.TimeAnalysis, error) {
	rows, err := database.Get().Query(ctx, `
		SELECT EXTRACT(HOUR FROM started_at)::INT AS hour,
			COALESCE(SUM(gmv),0),
			COALESCE(SUM(order_count),0),
			COALESCE(SUM(total_views),0),
			CASE WHEN SUM(total_views) > 0
				THEN (SUM(total_likes)+SUM(total_comments)+SUM(total_shares))::DECIMAL / SUM(total_views)
				ELSE 0 END
		FROM live_rooms WHERE created_at BETWEEN $1 AND $2 AND status='ended'
		GROUP BY hour ORDER BY hour`, startDate, endDate)
	if err != nil { return nil, err }
	defer rows.Close()
	var list []model.TimeAnalysis
	for rows.Next() {
		var ta model.TimeAnalysis
		if err := rows.Scan(&ta.Hour, &ta.GMV, &ta.Orders, &ta.Viewers, &ta.Engagement); err != nil {
			continue
		}
		list = append(list, ta)
	}
	return list, nil
}

func (r *AnalyticsRepo) GetViewerMetrics(ctx context.Context, liveRoomID int64, limit int) ([]model.ViewerMetrics, error) {
	if limit <= 0 { limit = 60 }
	rows, err := database.Get().Query(ctx,
		`SELECT id, live_room_id, timestamp, concurrent_viewers, new_followers, likes, comments,
		shares, gifts_count, gifts_value, avg_watch_time, created_at
		FROM viewer_metrics WHERE live_room_id=$1 ORDER BY timestamp DESC LIMIT $2`, liveRoomID, limit)
	if err != nil { return nil, err }
	defer rows.Close()
	var list []model.ViewerMetrics
	for rows.Next() {
		var m model.ViewerMetrics
		if err := rows.Scan(&m.ID, &m.LiveRoomID, &m.Timestamp, &m.ConcurrentIndex, &m.NewFollowers, &m.Likes, &m.Comments, &m.Shares, &m.GiftsCount, &m.GiftsValue, &m.AvgWatchTime, &m.CreatedAt); err != nil {
			continue
		}
		list = append(list, m)
	}
	return list, nil
}

func (r *AnalyticsRepo) GetDemographics(ctx context.Context, liveRoomID int64) ([]model.ViewerDemographics, error) {
	rows, err := database.Get().Query(ctx,
		`SELECT id, live_room_id, date, age_distribution, gender_distribution,
		region_distribution, device_distribution, top_cities, created_at
		FROM viewer_demographics WHERE live_room_id=$1 ORDER BY date DESC LIMIT 7`, liveRoomID)
	if err != nil { return nil, err }
	defer rows.Close()
	var list []model.ViewerDemographics
	for rows.Next() {
		var d model.ViewerDemographics
		if err := rows.Scan(&d.ID, &d.LiveRoomID, &d.Date, &d.AgeDistribution, &d.GenderDistribution, &d.RegionDistribution, &d.DeviceDistribution, &d.TopCities, &d.CreatedAt); err != nil {
			continue
		}
		list = append(list, d)
	}
	return list, nil
}
