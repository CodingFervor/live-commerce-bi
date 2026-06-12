package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/CodingFervor/live-commerce-bi/internal/cache"
	"github.com/CodingFervor/live-commerce-bi/internal/database"
	"github.com/CodingFervor/live-commerce-bi/internal/model"
	"github.com/CodingFervor/live-commerce-bi/pkg/logger"
)

// ═══ Intelligent Alert Engine ═══
// Evaluates alert rules periodically and triggers notifications
// Reference: Alibaba Cloud ARMS alerting architecture

type AlertEngine struct {
	notifier  *NotificationEngine
	mu        sync.Mutex
	cooldowns map[int64]time.Time // rule ID -> last triggered time
}

func NewAlertEngine() *AlertEngine {
	return &AlertEngine{
		notifier:  NewNotificationEngine(),
		cooldowns: make(map[int64]time.Time),
	}
}

// AlertEvaluationResult holds the result of evaluating a single rule
type AlertEvaluationResult struct {
	RuleID    int64   `json:"rule_id"`
	RuleName  string  `json:"rule_name"`
	Triggered bool    `json:"triggered"`
	Value     float64 `json:"value"`
	Threshold float64 `json:"threshold"`
	Message   string  `json:"message,omitempty"`
	Severity  string  `json:"severity"`
}

// EvaluateAll evaluates all enabled alert rules
func (e *AlertEngine) EvaluateAll(ctx context.Context) ([]AlertEvaluationResult, error) {
	rows, err := database.Get().Query(ctx, `
		SELECT id, name, metric, condition, threshold, timeframe, severity,
			notify_channels, notify_config, cooldown, is_enabled
		FROM alert_rules WHERE is_enabled = true`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []AlertEvaluationResult
	for rows.Next() {
		var rule model.AlertRule
		if err := rows.Scan(&rule.ID, &rule.Name, &rule.Metric, &rule.Condition,
			&rule.Threshold, &rule.Timeframe, &rule.Severity,
			&rule.NotifyChannels, &rule.NotifyConfig, &rule.Cooldown, &rule.IsEnabled); err != nil {
			continue
		}

		result := e.evaluateRule(ctx, &rule)
		results = append(results, result)

		if result.Triggered {
			go e.handleTriggeredAlert(ctx, &rule, &result)
		}
	}
	return results, nil
}

// EvaluateRule evaluates a single alert rule
func (e *AlertEngine) EvaluateRule(ctx context.Context, ruleID int64) (*AlertEvaluationResult, error) {
	var rule model.AlertRule
	err := database.Get().QueryRow(ctx, `
		SELECT id, name, metric, condition, threshold, timeframe, severity,
			notify_channels, notify_config, cooldown, is_enabled
		FROM alert_rules WHERE id=$1`, ruleID).Scan(
		&rule.ID, &rule.Name, &rule.Metric, &rule.Condition,
		&rule.Threshold, &rule.Timeframe, &rule.Severity,
		&rule.NotifyChannels, &rule.NotifyConfig, &rule.Cooldown, &rule.IsEnabled)
	if err != nil {
		return nil, err
	}

	result := e.evaluateRule(ctx, &rule)
	if result.Triggered {
		go e.handleTriggeredAlert(ctx, &rule, result)
	}
	return result, nil
}

func (e *AlertEngine) evaluateRule(ctx context.Context, rule *model.AlertRule) AlertEvaluationResult {
	result := AlertEvaluationResult{
		RuleID:    rule.ID,
		RuleName:  rule.Name,
		Threshold: rule.Threshold,
		Severity:  rule.Severity,
	}

	// Check cooldown
	e.mu.Lock()
	lastTriggered, hasCooldown := e.cooldowns[rule.ID]
	e.mu.Unlock()
	if hasCooldown && rule.Cooldown > 0 {
		cooldownDuration := time.Duration(rule.Cooldown) * time.Second
		if time.Since(lastTriggered) < cooldownDuration {
			result.Triggered = false
			return result
		}
	}

	// Get current metric value
	currentValue := e.getCurrentMetric(ctx, rule.Metric, rule.Timeframe)
	result.Value = currentValue

	// Evaluate condition
	triggered := false
	switch rule.Condition {
	case "gt":
		triggered = currentValue > rule.Threshold
	case "lt":
		triggered = currentValue < rule.Threshold
	case "eq":
		triggered = math.Abs(currentValue-rule.Threshold) < 0.01
	case "gte":
		triggered = currentValue >= rule.Threshold
	case "lte":
		triggered = currentValue <= rule.Threshold
	case "change_pct_gt":
		// Compare with previous period
		prevValue := e.getPreviousMetric(ctx, rule.Metric, rule.Timeframe)
		if prevValue > 0 {
			changePct := (currentValue - prevValue) / prevValue * 100
			triggered = changePct > rule.Threshold
			result.Value = changePct
		}
	case "change_pct_lt":
		prevValue := e.getPreviousMetric(ctx, rule.Metric, rule.Timeframe)
		if prevValue > 0 {
			changePct := (currentValue - prevValue) / prevValue * 100
			triggered = changePct < -rule.Threshold
			result.Value = changePct
		}
	}

	result.Triggered = triggered
	if triggered {
		result.Message = fmt.Sprintf("[%s] %s: 当前值 %.2f %s 阈值 %.2f",
			rule.Severity, rule.Name, result.Value,
			conditionSymbol(rule.Condition), rule.Threshold)
	}

	return result
}

func (e *AlertEngine) getCurrentMetric(ctx context.Context, metric, timeframe string) float64 {
	timeFilter := metricTimeFilter(timeframe, "lr")

	query := fmt.Sprintf(`
		SELECT COALESCE(SUM(gmv),0) FROM live_rooms lr WHERE 1=1 %s`, timeFilter)

	switch metric {
	case "gmv":
		query = fmt.Sprintf(`SELECT COALESCE(SUM(gmv),0) FROM live_rooms lr WHERE 1=1 %s`, timeFilter)
	case "viewers":
		query = fmt.Sprintf(`SELECT COALESCE(SUM(total_views),0) FROM live_rooms lr WHERE 1=1 %s`, timeFilter)
	case "order_count":
		query = fmt.Sprintf(`SELECT COALESCE(SUM(order_count),0) FROM live_rooms lr WHERE 1=1 %s`, timeFilter)
	case "conversion_rate":
		query = fmt.Sprintf(`SELECT COALESCE(AVG(conversion_rate),0) FROM live_rooms lr WHERE 1=1 %s`, timeFilter)
	case "peak_viewers":
		query = fmt.Sprintf(`SELECT COALESCE(MAX(peak_viewers),0) FROM live_rooms lr WHERE 1=1 %s`, timeFilter)
	case "avg_watch_time":
		query = fmt.Sprintf(`SELECT COALESCE(AVG(avg_viewers),0) FROM viewer_metrics vm WHERE 1=1 %s`,
			metricTimeFilter(timeframe, "vm"))
	case "engagement_rate":
		query = fmt.Sprintf(`SELECT COALESCE(AVG((total_likes+total_comments+total_shares))::DECIMAL /
			NULLIF(total_views,0), 0) FROM live_rooms lr WHERE 1=1 %s`, timeFilter)
	case "refund_rate":
		query = fmt.Sprintf(`SELECT COALESCE(
			COUNT(CASE WHEN status='refunded' THEN 1 END)::DECIMAL /
			NULLIF(COUNT(*),0), 0) FROM orders WHERE status NOT IN ('pending') %s`,
			metricTimeFilter(timeframe, ""))
	default:
		return 0
	}

	var value float64
	database.Get().QueryRow(ctx, query).Scan(&value)
	return value
}

func (e *AlertEngine) getPreviousMetric(ctx context.Context, metric, timeframe string) float64 {
	// Shift timeframe back by one period
	return e.getCurrentMetric(ctx, metric, "previous_"+timeframe)
}

func (e *AlertEngine) handleTriggeredAlert(ctx context.Context, rule *model.AlertRule, result *AlertEvaluationResult) {
	// Update cooldown
	e.mu.Lock()
	e.cooldowns[rule.ID] = time.Now()
	e.mu.Unlock()

	// Write alert history
	history := &model.AlertHistory{
		AlertRuleID:    rule.ID,
		Metric:         rule.Metric,
		TriggeredValue: result.Value,
		ThresholdValue: rule.Threshold,
		Message:        result.Message,
		Severity:       rule.Severity,
		Notified:       false,
	}

	err := database.Get().QueryRow(ctx, `
		INSERT INTO alert_history (alert_rule_id, metric, triggered_value, threshold_value, message, severity, notified)
		VALUES ($1,$2,$3,$4,$5,$6,false) RETURNING id, created_at`,
		history.AlertRuleID, history.Metric, history.TriggeredValue,
		history.ThresholdValue, history.Message, history.Severity,
	).Scan(&history.ID, &history.CreatedAt)

	if err != nil {
		logger.Error("Failed to write alert history: %v", err)
		return
	}

	// Send notifications
	if rule.NotifyChannels != "" && rule.NotifyConfig != "" {
		if err := e.notifier.Notify(ctx, rule.NotifyChannels, rule.NotifyConfig,
			"[BI告警] "+rule.Name, result.Message); err != nil {
			logger.Error("Alert notification failed for rule %d: %v", rule.ID, err)
		} else {
			// Mark as notified
			database.Get().Exec(ctx, "UPDATE alert_history SET notified=true WHERE id=$1", history.ID)
		}
	}

	// Push real-time alert via WebSocket
	alertData, _ := json.Marshal(map[string]interface{}{
		"type":      "alert_triggered",
		"rule_id":   rule.ID,
		"rule_name": rule.Name,
		"metric":    rule.Metric,
		"value":     result.Value,
		"threshold": rule.Threshold,
		"severity":  rule.Severity,
		"message":   result.Message,
	})
	PushRealtimeMetric(ctx, "live:room:alert", alertData)

	logger.Info("Alert triggered: rule=%s metric=%s value=%.2f threshold=%.2f",
		rule.Name, rule.Metric, result.Value, rule.Threshold)
}

// StartAlertScheduler starts the periodic alert evaluation loop
func (e *AlertEngine) StartAlertScheduler(ctx context.Context, interval time.Duration) {
	if interval == 0 {
		interval = 1 * time.Minute
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			results, err := e.EvaluateAll(ctx)
			if err != nil {
				logger.Error("Alert evaluation failed: %v", err)
				continue
			}
			triggered := 0
			for _, r := range results {
				if r.Triggered {
					triggered++
				}
			}
			if triggered > 0 {
				logger.Info("Alert evaluation: %d rules checked, %d triggered", len(results), triggered)
			}
		}
	}
}

func metricTimeFilter(timeframe, tableAlias string) string {
	prefix := ""
	if tableAlias != "" {
		prefix = tableAlias + "."
	}

	switch timeframe {
	case "1m":
		return fmt.Sprintf("AND %screated_at >= NOW() - INTERVAL '1 minute'", prefix)
	case "5m":
		return fmt.Sprintf("AND %screated_at >= NOW() - INTERVAL '5 minutes'", prefix)
	case "15m":
		return fmt.Sprintf("AND %screated_at >= NOW() - INTERVAL '15 minutes'", prefix)
	case "30m":
		return fmt.Sprintf("AND %screated_at >= NOW() - INTERVAL '30 minutes'", prefix)
	case "1h":
		return fmt.Sprintf("AND %screated_at >= NOW() - INTERVAL '1 hour'", prefix)
	case "6h":
		return fmt.Sprintf("AND %screated_at >= NOW() - INTERVAL '6 hours'", prefix)
	case "12h":
		return fmt.Sprintf("AND %screated_at >= NOW() - INTERVAL '12 hours'", prefix)
	case "24h":
		return fmt.Sprintf("AND %screated_at >= NOW() - INTERVAL '24 hours'", prefix)
	case "7d":
		return fmt.Sprintf("AND %screated_at >= NOW() - INTERVAL '7 days'", prefix)
	case "30d":
		return fmt.Sprintf("AND %screated_at >= NOW() - INTERVAL '30 days'", prefix)
	// Previous period variants for change comparison
	case "previous_1m":
		return fmt.Sprintf("AND %screated_at >= NOW() - INTERVAL '2 minutes' AND %screated_at < NOW() - INTERVAL '1 minute'", prefix, prefix)
	case "previous_5m":
		return fmt.Sprintf("AND %screated_at >= NOW() - INTERVAL '10 minutes' AND %screated_at < NOW() - INTERVAL '5 minutes'", prefix, prefix)
	case "previous_1h":
		return fmt.Sprintf("AND %screated_at >= NOW() - INTERVAL '2 hours' AND %screated_at < NOW() - INTERVAL '1 hour'", prefix, prefix)
	case "previous_24h":
		return fmt.Sprintf("AND %screated_at >= NOW() - INTERVAL '48 hours' AND %screated_at < NOW() - INTERVAL '24 hours'", prefix, prefix)
	case "previous_7d":
		return fmt.Sprintf("AND %screated_at >= NOW() - INTERVAL '14 days' AND %screated_at < NOW() - INTERVAL '7 days'", prefix, prefix)
	default:
		return fmt.Sprintf("AND %screated_at >= NOW() - INTERVAL '5 minutes'", prefix)
	}
}

func conditionSymbol(condition string) string {
	switch condition {
	case "gt":
		return ">"
	case "lt":
		return "<"
	case "eq":
		return "="
	case "gte":
		return ">="
	case "lte":
		return "<="
	default:
		return "vs"
	}
}
