package service

import (
	"context"

	"github.com/CodingFervor/live-commerce-bi/internal/model"
	"github.com/CodingFervor/live-commerce-bi/internal/repository"
)

type AnalyticsService struct {
	repo *repository.AnalyticsRepo
}

func NewAnalyticsService() *AnalyticsService {
	return &AnalyticsService{repo: repository.NewAnalyticsRepo()}
}

func (s *AnalyticsService) GetOverview(ctx context.Context, startDate, endDate string) (*model.OverviewStats, error) {
	return s.repo.GetOverviewStats(ctx, startDate, endDate)
}

func (s *AnalyticsService) GetGMVTrend(ctx context.Context, startDate, endDate string) ([]model.GMVAnalytics, error) {
	return s.repo.GetGMVTrend(ctx, startDate, endDate)
}

func (s *AnalyticsService) GetConversionFunnel(ctx context.Context, liveRoomID int64) ([]model.ConversionAnalysis, error) {
	return s.repo.GetConversionFunnel(ctx, liveRoomID)
}

func (s *AnalyticsService) GetPlatformComparison(ctx context.Context, startDate, endDate string) ([]model.PlatformComparison, error) {
	return s.repo.GetPlatformComparison(ctx, startDate, endDate)
}

func (s *AnalyticsService) GetCategoryAnalysis(ctx context.Context, startDate, endDate string) ([]model.CategoryAnalysis, error) {
	return s.repo.GetCategoryAnalysis(ctx, startDate, endDate)
}

func (s *AnalyticsService) GetTimeAnalysis(ctx context.Context, startDate, endDate string) ([]model.TimeAnalysis, error) {
	return s.repo.GetTimeAnalysis(ctx, startDate, endDate)
}

func (s *AnalyticsService) GetViewerMetrics(ctx context.Context, liveRoomID int64, limit int) ([]model.ViewerMetrics, error) {
	return s.repo.GetViewerMetrics(ctx, liveRoomID, limit)
}

func (s *AnalyticsService) GetDemographics(ctx context.Context, liveRoomID int64) ([]model.ViewerDemographics, error) {
	return s.repo.GetDemographics(ctx, liveRoomID)
}
