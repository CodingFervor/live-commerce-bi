package service

import (
	"context"
	"fmt"

	"github.com/CodingFervor/live-commerce-bi/internal/model"
	"github.com/CodingFervor/live-commerce-bi/internal/repository"
)

type AlertService struct {
	repo *repository.AlertRepo
}

func NewAlertService() *AlertService {
	return &AlertService{repo: repository.NewAlertRepo()}
}

func (s *AlertService) CreateRule(ctx context.Context, req *model.AlertRuleCreate, userID int64) (*model.AlertRule, error) {
	rule := &model.AlertRule{
		Name:           req.Name,
		Description:    req.Description,
		Metric:         req.Metric,
		Condition:      req.Condition,
		Threshold:      req.Threshold,
		Timeframe:      req.Timeframe,
		Severity:       req.Severity,
		NotifyChannels: req.NotifyChannels,
		NotifyConfig:   req.NotifyConfig,
		Cooldown:       req.Cooldown,
		IsEnabled:      true,
		CreatedBy:      userID,
	}
	if rule.Timeframe == "" { rule.Timeframe = "5m" }
	if rule.Severity == "" { rule.Severity = "warning" }
	if rule.NotifyChannels == "" { rule.NotifyChannels = `["email"]` }
	if rule.NotifyConfig == "" { rule.NotifyConfig = "{}" }
	if rule.Cooldown == 0 { rule.Cooldown = 300 }
	if err := s.repo.CreateRule(ctx, rule); err != nil {
		return nil, err
	}
	return rule, nil
}

func (s *AlertService) ListRules(ctx context.Context) ([]model.AlertRule, error) {
	return s.repo.ListRules(ctx)
}

func (s *AlertService) GetRule(ctx context.Context, id int64) (*model.AlertRule, error) {
	return s.repo.GetRule(ctx, id)
}

func (s *AlertService) UpdateRule(ctx context.Context, id int64, req *model.AlertRuleCreate) (*model.AlertRule, error) {
	rule, err := s.repo.GetRule(ctx, id)
	if err != nil { return nil, err }
	rule.Name = req.Name
	rule.Description = req.Description
	rule.Metric = req.Metric
	rule.Condition = req.Condition
	rule.Threshold = req.Threshold
	if req.Timeframe != "" { rule.Timeframe = req.Timeframe }
	if req.Severity != "" { rule.Severity = req.Severity }
	if req.NotifyChannels != "" { rule.NotifyChannels = req.NotifyChannels }
	if req.NotifyConfig != "" { rule.NotifyConfig = req.NotifyConfig }
	if err := s.repo.UpdateRule(ctx, rule); err != nil {
		return nil, err
	}
	return rule, nil
}

func (s *AlertService) DeleteRule(ctx context.Context, id int64) error {
	return s.repo.DeleteRule(ctx, id)
}

func (s *AlertService) GetHistory(ctx context.Context, ruleID int64, page, pageSize int) ([]model.AlertHistory, int, error) {
	return s.repo.ListHistory(ctx, ruleID, page, pageSize)
}

func (s *AlertService) TestAlert(ctx context.Context, id int64) error {
	rule, err := s.repo.GetRule(ctx, id)
	if err != nil { return err }
	// Create test alert history entry
	history := &model.AlertHistory{
		AlertRuleID:    rule.ID,
		Metric:         rule.Metric,
		TriggeredValue: rule.Threshold,
		ThresholdValue: rule.Threshold,
		Message:        fmt.Sprintf("[测试] %s 触发测试", rule.Name),
		Severity:       rule.Severity,
		Notified:       false,
	}
	return s.repo.CreateHistory(ctx, history)
}
