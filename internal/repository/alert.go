package repository

import (
	"context"

	"github.com/CodingFervor/live-commerce-bi/internal/database"
	"github.com/CodingFervor/live-commerce-bi/internal/model"
)

type AlertRepo struct{}

func NewAlertRepo() *AlertRepo { return &AlertRepo{} }

func (r *AlertRepo) CreateRule(ctx context.Context, rule *model.AlertRule) error {
	return database.Get().QueryRow(ctx,
		`INSERT INTO alert_rules (name, description, metric, condition, threshold, timeframe, severity,
		notify_channels, notify_config, cooldown, is_enabled, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12) RETURNING id, created_at`,
		rule.Name, rule.Description, rule.Metric, rule.Condition, rule.Threshold, rule.Timeframe, rule.Severity,
		rule.NotifyChannels, rule.NotifyConfig, rule.Cooldown, rule.IsEnabled, rule.CreatedBy,
	).Scan(&rule.ID, &rule.CreatedAt)
}

func (r *AlertRepo) ListRules(ctx context.Context) ([]model.AlertRule, error) {
	rows, err := database.Get().Query(ctx,
		`SELECT id, name, description, metric, condition, threshold, timeframe, severity,
		notify_channels, notify_config, cooldown, is_enabled, created_by, created_at, updated_at
		FROM alert_rules ORDER BY created_at DESC`)
	if err != nil { return nil, err }
	defer rows.Close()
	var list []model.AlertRule
	for rows.Next() {
		var rule model.AlertRule
		if err := rows.Scan(&rule.ID, &rule.Name, &rule.Description, &rule.Metric, &rule.Condition,
			&rule.Threshold, &rule.Timeframe, &rule.Severity, &rule.NotifyChannels, &rule.NotifyConfig,
			&rule.Cooldown, &rule.IsEnabled, &rule.CreatedBy, &rule.CreatedAt, &rule.UpdatedAt); err != nil {
			continue
		}
		list = append(list, rule)
	}
	return list, nil
}

func (r *AlertRepo) GetRule(ctx context.Context, id int64) (*model.AlertRule, error) {
	var rule model.AlertRule
	err := database.Get().QueryRow(ctx,
		`SELECT id, name, description, metric, condition, threshold, timeframe, severity,
		notify_channels, notify_config, cooldown, is_enabled, created_by, created_at, updated_at
		FROM alert_rules WHERE id=$1`, id).Scan(
		&rule.ID, &rule.Name, &rule.Description, &rule.Metric, &rule.Condition,
		&rule.Threshold, &rule.Timeframe, &rule.Severity, &rule.NotifyChannels, &rule.NotifyConfig,
		&rule.Cooldown, &rule.IsEnabled, &rule.CreatedBy, &rule.CreatedAt, &rule.UpdatedAt)
	return &rule, err
}

func (r *AlertRepo) UpdateRule(ctx context.Context, rule *model.AlertRule) error {
	_, err := database.Get().Exec(ctx,
		`UPDATE alert_rules SET name=$1, description=$2, metric=$3, condition=$4, threshold=$5,
		timeframe=$6, severity=$7, notify_channels=$8, notify_config=$9, cooldown=$10, is_enabled=$11 WHERE id=$12`,
		rule.Name, rule.Description, rule.Metric, rule.Condition, rule.Threshold,
		rule.Timeframe, rule.Severity, rule.NotifyChannels, rule.NotifyConfig, rule.Cooldown, rule.IsEnabled, rule.ID)
	return err
}

func (r *AlertRepo) DeleteRule(ctx context.Context, id int64) error {
	_, err := database.Get().Exec(ctx, "DELETE FROM alert_rules WHERE id=$1", id)
	return err
}

func (r *AlertRepo) CreateHistory(ctx context.Context, h *model.AlertHistory) error {
	return database.Get().QueryRow(ctx,
		`INSERT INTO alert_history (alert_rule_id, metric, triggered_value, threshold_value, message, severity, notified)
		VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id, created_at`,
		h.AlertRuleID, h.Metric, h.TriggeredValue, h.ThresholdValue, h.Message, h.Severity, h.Notified,
	).Scan(&h.ID, &h.CreatedAt)
}

func (r *AlertRepo) ListHistory(ctx context.Context, ruleID int64, page, pageSize int) ([]model.AlertHistory, int, error) {
	var total int
	database.Get().QueryRow(ctx, "SELECT COUNT(*) FROM alert_history WHERE alert_rule_id=$1", ruleID).Scan(&total)
	offset := (page - 1) * pageSize
	rows, err := database.Get().Query(ctx,
		`SELECT ah.id, ah.alert_rule_id, ah.metric, ah.triggered_value, ah.threshold_value, ah.message,
		ah.severity, ah.notified, ah.acknowledged, ah.acknowledged_by, ah.acknowledged_at,
		ah.resolved_at, ah.created_at, ar.name AS rule_name
		FROM alert_history ah JOIN alert_rules ar ON ah.alert_rule_id = ar.id
		WHERE ah.alert_rule_id=$1 ORDER BY ah.created_at DESC LIMIT $2 OFFSET $3`, ruleID, pageSize, offset)
	if err != nil { return nil, 0, err }
	defer rows.Close()
	var list []model.AlertHistory
	for rows.Next() {
		var h model.AlertHistory
		if err := rows.Scan(&h.ID, &h.AlertRuleID, &h.Metric, &h.TriggeredValue, &h.ThresholdValue, &h.Message,
			&h.Severity, &h.Notified, &h.Acknowledged, &h.AcknowledgedBy, &h.AcknowledgedAt,
			&h.ResolvedAt, &h.CreatedAt, &h.RuleName); err != nil {
			continue
		}
		list = append(list, h)
	}
	return list, total, nil
}

func (r *AlertRepo) GetEnabledRules(ctx context.Context) ([]model.AlertRule, error) {
	rows, err := database.Get().Query(ctx,
		`SELECT id, name, metric, condition, threshold, timeframe, severity, notify_channels, notify_config
		FROM alert_rules WHERE is_enabled=true`)
	if err != nil { return nil, err }
	defer rows.Close()
	var list []model.AlertRule
	for rows.Next() {
		var rule model.AlertRule
		if err := rows.Scan(&rule.ID, &rule.Name, &rule.Metric, &rule.Condition, &rule.Threshold,
			&rule.Timeframe, &rule.Severity, &rule.NotifyChannels, &rule.NotifyConfig); err != nil {
			continue
		}
		list = append(list, rule)
	}
	return list, nil
}
