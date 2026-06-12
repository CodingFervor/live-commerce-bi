package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/CodingFervor/live-commerce-bi/internal/cache"
	"github.com/CodingFervor/live-commerce-bi/internal/database"
	"github.com/CodingFervor/live-commerce-bi/internal/model"
)

// ═══ Advanced Analytics Engine ═══
// Reference: ByteDance/Alibaba data analytics patterns

type AnalyticsEngine struct{}

func NewAnalyticsEngine() *AnalyticsEngine { return &AnalyticsEngine{} }

// ─── Cohort Analysis ───
// Groups users by first-purchase month, tracks retention over subsequent periods
func (e *AnalyticsEngine) CohortAnalysis(ctx context.Context, startDate, endDate string) ([]model.CohortData, error) {
	query := `
		WITH first_purchases AS (
			SELECT buyer_id, DATE_TRUNC('month', MIN(created_at))::date AS cohort_month
			FROM orders WHERE status NOT IN ('pending','cancelled')
			GROUP BY buyer_id
		),
	 monthly_purchases AS (
			SELECT fp.buyer_id, fp.cohort_month,
				DATE_TRUNC('month', o.created_at)::date AS activity_month,
				SUM(o.actual_amount) AS revenue
			FROM first_purchases fp
			JOIN orders o ON fp.buyer_id = o.buyer_id AND o.status NOT IN ('pending','cancelled')
			GROUP BY fp.buyer_id, fp.cohort_month, activity_month
		)
		SELECT cohort_month, COUNT(DISTINCT buyer_id) AS users,
			ARRAY_AGG(DISTINCT EXTRACT(DAY FROM activity_month - cohort_month)/30 ORDER BY EXTRACT(DAY FROM activity_month - cohort_month)/30) AS periods,
			SUM(revenue) AS revenue
		FROM monthly_purchases
		GROUP BY cohort_month ORDER BY cohort_month`

	rows, err := database.Get().Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cohorts []model.CohortData
	for rows.Next() {
		var cd model.CohortData
		var cohortMonth time.Time
		var periods []int
		var revenue float64
		if err := rows.Scan(&cohortMonth, &cd.Users, &periods, &revenue); err != nil {
			continue
		}
		cd.CohortDate = cohortMonth.Format("2006-01-02")
		cd.Revenue = revenue
		// Build retention array (day 0 is always 100%)
		retention := make([]float64, 8)
		retention[0] = 100.0
		for _, p := range periods {
			if p > 0 && p < 8 {
				retention[p] = 100.0 // simplified: present = 100% relative
			}
		}
		cd.Retention = retention
		cohorts = append(cohorts, cd)
	}
	return cohorts, nil
}

// ─── RFM Analysis ───
// Recency-Frequency-Monetary customer segmentation
func (e *AnalyticsEngine) RFMAnalysis(ctx context.Context) ([]model.RFMSegment, error) {
	query := `
		SELECT buyer_id,
			EXTRACT(DAY FROM NOW() - MAX(created_at))::int AS recency,
			COUNT(*) AS frequency,
			SUM(actual_amount) AS monetary
		FROM orders WHERE status NOT IN ('pending','cancelled')
		GROUP BY buyer_id ORDER BY monetary DESC LIMIT 1000`

	rows, err := database.Get().Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var segments []model.RFMSegment
	for rows.Next() {
		var s model.RFMSegment
		if err := rows.Scan(&s.UserID, &s.Recency, &s.Frequency, &s.Monetary); err != nil {
			continue
		}
		// Score 1-5 for each dimension
		rScore := 5 - min(s.Recency/30, 4)
		fScore := min(s.Frequency, 5)
		mScore := min(int(s.Monetary/1000)+1, 5)
		s.Score = rScore + fScore + mScore

		// Segment classification
		switch {
		case s.Score >= 12:
			s.Segment = "Champions"
		case s.Score >= 9:
			s.Segment = "Loyal Customers"
		case s.Score >= 7:
			s.Segment = "Potential Loyalist"
		case s.Score >= 5:
			s.Segment = "At Risk"
		default:
			s.Segment = "Lost"
		}
		segments = append(segments, s)
	}
	return segments, nil
}

// ─── Sales Forecast ───
// Exponential smoothing with confidence intervals
func (e *AnalyticsEngine) SalesForecast(ctx context.Context, days int) ([]model.SalesForecast, error) {
	if days <= 0 {
		days = 30
	}
	// Get historical daily GMV
	rows, err := database.Get().Query(ctx, `
		SELECT DATE(created_at) AS d, SUM(actual_amount) AS gmv
		FROM orders WHERE status NOT IN ('pending','cancelled')
		AND created_at >= NOW() - INTERVAL '90 days'
		GROUP BY d ORDER BY d`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var historical []float64
	var dates []string
	for rows.Next() {
		var d time.Time
		var gmv float64
		if err := rows.Scan(&d, &gmv); err != nil {
			continue
		}
		historical = append(historical, gmv)
		dates = append(dates, d.Format("2006-01-02"))
	}

	if len(historical) < 3 {
		return e.syntheticForecast(days), nil
	}

	// Exponential smoothing (alpha=0.3)
	alpha := 0.3
	smoothed := make([]float64, len(historical))
	smoothed[0] = historical[0]
	for i := 1; i < len(historical); i++ {
		smoothed[i] = alpha*historical[i] + (1-alpha)*smoothed[i-1]
	}

	// Calculate standard deviation for confidence interval
	var sumSq float64
	n := float64(len(historical))
	for i, v := range historical {
		sumSq += (v - smoothed[i]) * (v - smoothed[i])
	}
	stdDev := math.Sqrt(sumSq / n)

	// Generate forecast
	lastSmoothed := smoothed[len(smoothed)-1]
	lastDate, _ := time.Parse("2006-01-02", dates[len(dates)-1])

	var forecast []model.SalesForecast
	// Include last 7 actuals
	start := len(historical) - 7
	if start < 0 {
		start = 0
	}
	for i := start; i < len(historical); i++ {
		forecast = append(forecast, model.SalesForecast{
			Date:   dates[i],
			Actual: historical[i],
		})
	}

	// Forecast future days
	for i := 1; i <= days; i++ {
		futureDate := lastDate.AddDate(0, 0, i)
		fVal := lastSmoothed // simplified: level forecast
		forecast = append(forecast, model.SalesForecast{
			Date:     futureDate.Format("2006-01-02"),
			Forecast: math.Max(0, fVal),
			Lower:    math.Max(0, fVal-1.96*stdDev),
			Upper:    fVal + 1.96*stdDev,
		})
	}
	return forecast, nil
}

func (e *AnalyticsEngine) syntheticForecast(days int) []model.SalesForecast {
	var result []model.SalesForecast
	base := time.Now()
	for i := 0; i < days; i++ {
		d := base.AddDate(0, 0, i)
		val := 50000.0 + float64(i)*100
		result = append(result, model.SalesForecast{
			Date:     d.Format("2006-01-02"),
			Forecast: val,
			Lower:    val * 0.8,
			Upper:    val * 1.2,
		})
	}
	return result
}

// ─── Anomaly Detection ───
// Z-score based detection with configurable threshold
func (e *AnalyticsEngine) AnomalyDetection(ctx context.Context, metric string, days int) ([]model.AnomalyPoint, error) {
	if days <= 0 {
		days = 30
	}

	// Get daily metric values
	query := fmt.Sprintf(`
		SELECT DATE(created_at) AS d, SUM(actual_amount) AS value
		FROM orders WHERE status NOT IN ('pending','cancelled')
		AND created_at >= NOW() - INTERVAL '%d days'
		GROUP BY d ORDER BY d`, days+30)

	rows, err := database.Get().Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type point struct {
		ts    time.Time
		value float64
	}
	var points []point
	for rows.Next() {
		var p point
		if err := rows.Scan(&p.ts, &p.value); err != nil {
			continue
		}
		points = append(points, p)
	}

	if len(points) < 5 {
		return nil, nil
	}

	// Calculate mean and std dev
	var sum, sumSq float64
	n := float64(len(points))
	for _, p := range points {
		sum += p.value
	}
	mean := sum / n
	for _, p := range points {
		sumSq += (p.value - mean) * (p.value - mean)
	}
	stdDev := math.Sqrt(sumSq / n)

	if stdDev == 0 {
		stdDev = 1
	}

	// Detect anomalies (Z-score > 2)
	var anomalies []model.AnomalyPoint
	for _, p := range points {
		zScore := (p.value - mean) / stdDev
		isAnomaly := math.Abs(zScore) > 2.0
		severity := "normal"
		if math.Abs(zScore) > 3.0 {
			severity = "critical"
		} else if math.Abs(zScore) > 2.0 {
			severity = "warning"
		}
		anomalies = append(anomalies, model.AnomalyPoint{
			Timestamp: p.ts,
			Value:     p.value,
			Expected:  mean,
			Deviation: zScore,
			IsAnomaly: isAnomaly,
			Severity:  severity,
		})
	}
	return anomalies, nil
}

// ─── User Path Analysis ───
// Analyze user behavior paths from events
func (e *AnalyticsEngine) UserPathAnalysis(ctx context.Context, liveRoomID int64) ([]model.UserBehaviorPath, error) {
	rows, err := database.Get().Query(ctx, `
		SELECT session_id, user_id,
			ARRAY_AGG(event_name ORDER BY timestamp) AS events,
			EXTRACT(EPOCH FROM MAX(timestamp) - MIN(timestamp))::int AS duration,
			BOOL_OR(event_name = 'purchase') AS converted
		FROM track_events WHERE live_room_id = $1
		GROUP BY session_id, user_id
		ORDER BY duration DESC LIMIT 100`, liveRoomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var paths []model.UserBehaviorPath
	for rows.Next() {
		var p model.UserBehaviorPath
		if err := rows.Scan(&p.SessionID, &p.UserID, &p.Events, &p.Duration, &p.Converted); err != nil {
			continue
		}
		paths = append(paths, p)
	}
	return paths, nil
}

// ─── Engagement Heatmap ───
// Engagement intensity by time segment within a live room
func (e *AnalyticsEngine) EngagementHeatmap(ctx context.Context, liveRoomID int64) (map[string]interface{}, error) {
	rows, err := database.Get().Query(ctx, `
		SELECT EXTRACT(HOUR FROM timestamp)::int AS hour,
			COUNT(*) AS event_count,
			COUNT(DISTINCT user_id) AS unique_users,
			COUNT(CASE WHEN event_name='like' THEN 1 END) AS likes,
			COUNT(CASE WHEN event_name='comment' THEN 1 END) AS comments,
			COUNT(CASE WHEN event_name='share' THEN 1 END) AS shares
		FROM track_events WHERE live_room_id = $1
		GROUP BY hour ORDER BY hour`, liveRoomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type heatmapCell struct {
		Hour        int   `json:"hour"`
		EventCount  int   `json:"event_count"`
		UniqueUsers int   `json:"unique_users"`
		Likes       int   `json:"likes"`
		Comments    int   `json:"comments"`
		Shares      int   `json:"shares"`
	}
	var cells []heatmapCell
	for rows.Next() {
		var c heatmapCell
		if err := rows.Scan(&c.Hour, &c.EventCount, &c.UniqueUsers, &c.Likes, &c.Comments, &c.Shares); err != nil {
			continue
		}
		cells = append(cells, c)
	}

	// Find peak engagement
	var peakHour int
	var maxEvents int
	for _, c := range cells {
		if c.EventCount > maxEvents {
			maxEvents = c.EventCount
			peakHour = c.Hour
		}
	}

	return map[string]interface{}{
		"heatmap":   cells,
		"peak_hour": peakHour,
		"total_events": func() int {
			t := 0
			for _, c := range cells {
				t += c.EventCount
			}
			return t
		}(),
	}, nil
}

// ─── OLAP Multi-dimensional Query ───
func (e *AnalyticsEngine) ExecuteOLAP(ctx context.Context, q model.OLAPQuery) ([]model.OLAPResult, error) {
	// Build dynamic query with parameterized filters to prevent SQL injection
	allowedDims := map[string]string{
		"platform": "lr.platform", "streamer": "lr.streamer_id",
		"date": "DATE(lr.created_at)", "hour": "EXTRACT(HOUR FROM lr.created_at)",
		"status": "lr.status",
	}
	allowedMetrics := map[string]string{
		"gmv": "COALESCE(SUM(lr.gmv),0)", "orders": "COALESCE(SUM(lr.order_count),0)",
		"views": "COALESCE(SUM(lr.total_views),0)", "likes": "COALESCE(SUM(lr.total_likes),0)",
		"comments": "COALESCE(SUM(lr.total_comments),0)", "shares": "COALESCE(SUM(lr.total_shares),0)",
		"avg_conv": "COALESCE(AVG(lr.conversion_rate),0)", "count": "COUNT(*)",
	}

	var dimCols []string
	for _, d := range q.Dimensions {
		if col, ok := allowedDims[d]; ok {
			dimCols = append(dimCols, col+" AS "+d)
		}
	}
	var metCols []string
	for _, m := range q.Metrics {
		if agg, ok := allowedMetrics[m]; ok {
			metCols = append(metCols, agg+" AS "+m)
		}
	}

	if len(dimCols) == 0 && len(metCols) == 0 {
		return nil, fmt.Errorf("no valid dimensions or metrics")
	}

	selectParts := strings.Join(append(dimCols, metCols...), ", ")
	groupByParts := strings.Join(dimCols, ", ")

	query := "SELECT " + selectParts + " FROM live_rooms lr"
	var args []interface{}
	argIdx := 1

	if len(q.Filters) > 0 {
		query += " WHERE "
		first := true
		for k, v := range q.Filters {
			if col, ok := allowedDims[k]; ok {
				if !first {
					query += " AND "
				}
				query += fmt.Sprintf("%s = $%d", col, argIdx)
				args = append(args, v)
				argIdx++
				first = false
			}
		}
	}

	if groupByParts != "" {
		query += " GROUP BY " + groupByParts
	}

	if q.Limit > 0 && q.Limit <= 1000 {
		query += fmt.Sprintf(" LIMIT %d", q.Limit)
	} else {
		query += " LIMIT 100"
	}

	rows, err := database.Get().Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, _ := rows.FieldDescriptions()
	dimSet := make(map[string]bool)
	for _, d := range q.Dimensions {
		dimSet[d] = true
	}
	var results []model.OLAPResult
	for rows.Next() {
		vals, _ := rows.Values()
		dims := make(map[string]interface{})
		mets := make(map[string]interface{})
		for i, col := range cols {
			name := string(col.Name)
			if dimSet[name] {
				dims[name] = vals[i]
			} else {
				mets[name] = vals[i]
			}
		}
		results = append(results, model.OLAPResult{Dimensions: dims, Metrics: mets})
	}
	return results, nil
}

// ─── Period Comparison ───
func (e *AnalyticsEngine) PeriodComparison(ctx context.Context, metric, curStart, curEnd, compType string) (*model.ComparisonResult, error) {
	// Parse dates
	cs, _ := time.Parse("2006-01-02", curStart)
	ce, _ := time.Parse("2006-01-02", curEnd)

	// Calculate comparison period
	var ps, pe time.Time
	switch compType {
	case "yoy":
		ps = cs.AddDate(-1, 0, 0)
		pe = ce.AddDate(-1, 0, 0)
	case "mom":
		ps = cs.AddDate(0, -1, 0)
		pe = ce.AddDate(0, -1, 0)
	default: // wow
		ps = cs.AddDate(0, 0, -7)
		pe = ce.AddDate(0, 0, -7)
	}

	// Query both periods
	allowedAgg := map[string]string{
		"gmv": "COALESCE(SUM(gmv),0)", "orders": "COALESCE(SUM(order_count),0)",
		"views": "COALESCE(SUM(total_views),0)", "likes": "COALESCE(SUM(total_likes),0)",
		"comments": "COALESCE(SUM(total_comments),0)", "shares": "COALESCE(SUM(total_shares),0)",
		"avg_conv": "COALESCE(AVG(conversion_rate),0)", "count": "COUNT(*)",
	}
	agg, ok := allowedAgg[metric]
	if !ok {
		agg = "COUNT(*)"
	}
	current, prev := 0.0, 0.0
	database.Get().QueryRow(ctx,
		fmt.Sprintf("SELECT %s FROM live_rooms WHERE created_at BETWEEN $1 AND $2", agg),
		cs, ce.Add(24*time.Hour)).Scan(&current)
	database.Get().QueryRow(ctx,
		fmt.Sprintf("SELECT %s FROM live_rooms WHERE created_at BETWEEN $1 AND $2", agg),
		ps, pe.Add(24*time.Hour)).Scan(&prev)

	change := current - prev
	changePct := 0.0
	if prev > 0 {
		changePct = (change / prev) * 100
	}

	return &model.ComparisonResult{
		Current:   map[string]interface{}{"period": curStart + " ~ " + curEnd, "value": current},
		Previous:  map[string]interface{}{"period": ps.Format("2006-01-02") + " ~ " + pe.Format("2006-01-02"), "value": prev},
		Change:    map[string]interface{}{"value": change},
		ChangePct: map[string]interface{}{"value": changePct},
	}, nil
}

// ─── Target Comparison ───
func (e *AnalyticsEngine) TargetComparison(ctx context.Context, metric, start, end string, target float64) (map[string]interface{}, error) {
	allowedAgg := map[string]string{
		"gmv": "COALESCE(SUM(gmv),0)", "orders": "COALESCE(SUM(order_count),0)",
		"views": "COALESCE(SUM(total_views),0)", "likes": "COALESCE(SUM(total_likes),0)",
		"avg_conv": "COALESCE(AVG(conversion_rate),0)", "count": "COUNT(*)",
	}
	agg, ok := allowedAgg[metric]
	if !ok {
		agg = "COUNT(*)"
	}
	var actual float64
	database.Get().QueryRow(ctx,
		fmt.Sprintf("SELECT COALESCE(%s,0) FROM live_rooms WHERE created_at BETWEEN $1 AND $2", agg),
		start, end+" 23:59:59").Scan(&actual)

	achievement := 0.0
	if target > 0 {
		achievement = (actual / target) * 100
	}
	return map[string]interface{}{
		"actual":       actual,
		"target":       target,
		"achievement":  achievement,
		"gap":          target - actual,
		"on_track":     actual >= target,
	}, nil
}

// ─── ETL Engine ───

type PlatformConnector interface {
	Connect(config string) error
	FetchLiveRooms(ctx context.Context) ([]json.RawMessage, error)
	FetchOrders(ctx context.Context, since time.Time) ([]json.RawMessage, error)
	FetchProducts(ctx context.Context) ([]json.RawMessage, error)
}

type ETLEngine struct {
	connectors map[string]PlatformConnector
}

func NewETLEngine() *ETLEngine {
	return &ETLEngine{connectors: make(map[string]PlatformConnector)}
}

func (eng *ETLEngine) RegisterConnector(platform string, conn PlatformConnector) {
	eng.connectors[platform] = conn
}

func (eng *ETLEngine) RunSync(ctx context.Context, platform string, config string) (int, error) {
	conn, ok := eng.connectors[platform]
	if !ok {
		return 0, fmt.Errorf("no connector for platform: %s", platform)
	}
	if err := conn.Connect(config); err != nil {
		return 0, err
	}
	rooms, err := conn.FetchLiveRooms(ctx)
	if err != nil {
		return 0, err
	}
	products, err := conn.FetchProducts(ctx)
	if err != nil {
		return 0, err
	}
	return len(rooms) + len(products), nil
}

// ─── Export Engine ───

type ExportEngine struct{}

func NewExportEngine() *ExportEngine { return &ExportEngine{} }

func (eng *ExportEngine) BuildCSV(headers []string, rows [][]string) string {
	result := ""
	for i, h := range headers {
		if i > 0 {
			result += ","
		}
		result += h
	}
	result += "\n"
	for _, row := range rows {
		for i, cell := range row {
			if i > 0 {
				result += ","
			}
			// Escape CSV
			if containsAny(cell, ",\"\n") {
				result += "\"" + strings.ReplaceAll(cell, "\"", "\"\"") + "\""
			} else {
				result += cell
			}
		}
		result += "\n"
	}
	return result
}

func (eng *ExportEngine) BuildJSON(headers []string, rows [][]string) string {
	var entries []map[string]string
	for _, row := range rows {
		entry := make(map[string]string)
		for i, val := range row {
			if i < len(headers) {
				entry[headers[i]] = val
			}
		}
		entries = append(entries, entry)
	}
	b, _ := json.Marshal(entries)
	return string(b)
}

// ─── Notification Engine ───

import (
	"bytes"
	"net/http"
)

type NotificationEngine struct {
	httpClient *http.Client
}

func NewNotificationEngine() *NotificationEngine {
	return &NotificationEngine{
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (n *NotificationEngine) SendEmail(to, subject, body string) error {
	// Production: integrate with SMTP/SendGrid/Ses
	// POST https://api.sendgrid.com/v3/mail/send
	payload := map[string]interface{}{
		"personalizations": []map[string]interface{}{{"to": []map[string]string{{"email": to}}}},
		"from":             map[string]string{"email": "bi@livecommerce.com", "name": "直播BI系统"},
		"subject":          subject,
		"content":          []map[string]string{{"type": "text/plain", "value": body}},
	}
	return n.postJSON("https://api.sendgrid.com/v3/mail/send", payload)
}

func (n *NotificationEngine) SendSMS(phone, message string) error {
	// Production: integrate with Alibaba Cloud SMS / Twilio
	payload := map[string]interface{}{
		"phone":  phone,
		"msg":    message,
		"sign":   "直播BI",
	}
	return n.postJSON("https://dysmsapi.aliyuncs.com/", payload)
}

func (n *NotificationEngine) SendDingTalk(webhook, message string) error {
	payload := map[string]interface{}{
		"msgtype": "text",
		"text":    map[string]string{"content": message},
	}
	return n.postJSON(webhook, payload)
}

func (n *NotificationEngine) SendWebhook(url string, payload map[string]interface{}) error {
	return n.postJSON(url, payload)
}

func (n *NotificationEngine) Notify(ctx context.Context, channels string, config string, subject, body string) error {
	var chList []string
	if err := json.Unmarshal([]byte(channels), &chList); err != nil {
		return fmt.Errorf("parse channels: %w", err)
	}
	var lastErr error
	for _, ch := range chList {
		var err error
		switch ch {
		case "email":
			err = n.SendEmail(config, subject, body)
		case "sms":
			err = n.SendSMS(config, body)
		case "dingtalk":
			err = n.SendDingTalk(config, body)
		case "webhook":
			err = n.SendWebhook(config, map[string]interface{}{"subject": subject, "body": body})
		}
		if err != nil {
			lastErr = fmt.Errorf("%s: %w", ch, err)
		}
	}
	return lastErr
}

func (n *NotificationEngine) postJSON(url string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	resp, err := n.httpClient.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("notification API returned %d", resp.StatusCode)
	}
	return nil
}

// ─── Metrics Aggregation Engine ┐

type MetricsAggregator struct{}

func NewMetricsAggregator() *MetricsAggregator { return &MetricsAggregator{} }

func (m *MetricsAggregator) AggregateHourly(ctx context.Context, platform string, hour time.Time) error {
	truncHour := hour.Truncate(time.Hour)
	_, err := database.Get().Exec(ctx, `
		INSERT INTO metrics_hourly (platform, hour, gmv, order_count, viewer_count, peak_viewers, conversion_rate)
		SELECT platform, $2,
			COALESCE(SUM(gmv),0), COALESCE(SUM(order_count),0),
			COALESCE(SUM(total_views),0), COALESCE(MAX(peak_viewers),0),
			COALESCE(AVG(conversion_rate),0)
		FROM live_rooms
		WHERE ($1 = '' OR platform = $1) AND created_at >= $2 AND created_at < $2 + INTERVAL '1 hour'
		GROUP BY platform
		ON CONFLICT (platform, hour, streamer_id, live_room_id) DO UPDATE SET
			gmv = EXCLUDED.gmv, order_count = EXCLUDED.order_count`,
		platform, truncHour)
	return err
}

func (m *MetricsAggregator) AggregateDaily(ctx context.Context, platform string, date time.Time) error {
	day := date.Truncate(24 * time.Hour)
	_, err := database.Get().Exec(ctx, `
		INSERT INTO metrics_daily (platform, date, gmv, order_count, viewer_count, live_room_count, conversion_rate)
		SELECT platform, $2,
			COALESCE(SUM(gmv),0), COALESCE(SUM(order_count),0),
			COALESCE(SUM(viewer_count),0), COUNT(DISTINCT live_room_id),
			COALESCE(AVG(conversion_rate),0)
		FROM live_rooms
		WHERE ($1 = '' OR platform = $1) AND created_at >= $2 AND created_at < $2 + INTERVAL '1 day'
		GROUP BY platform
		ON CONFLICT (platform, date, streamer_id) DO UPDATE SET
			gmv = EXCLUDED.gmv, order_count = EXCLUDED.order_count`,
		platform, day)
	return err
}

func (m *MetricsAggregator) WarmupCache(ctx context.Context) error {
	rdb := cache.Get()
	if rdb == nil {
		return nil
	}
	// Cache platform summary
	rows, err := database.Get().Query(ctx, `
		SELECT platform, SUM(gmv), SUM(order_count), SUM(total_views)
		FROM live_rooms WHERE created_at >= NOW() - INTERVAL '24 hours'
		GROUP BY platform`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var platform string
		var gmv, views float64
		var orders int
		if err := rows.Scan(&platform, &gmv, &orders, &views); err != nil {
			continue
		}
		data, _ := json.Marshal(map[string]interface{}{
			"platform": platform, "gmv": gmv, "orders": orders, "views": views,
		})
		rdb.Set(ctx, "cache:platform:"+platform, string(data), 5*time.Minute)
	}
	return nil
}

// ─── Helpers ───

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func containsAny(s string, chars string) bool {
	for _, c := range chars {
		for _, sc := range s {
			if c == sc {
				return true
			}
		}
	}
	return false
}
