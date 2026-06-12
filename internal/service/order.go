package service

import (
	"context"

	"github.com/CodingFervor/live-commerce-bi/internal/model"
	"github.com/CodingFervor/live-commerce-bi/internal/repository"
)

type OrderService struct {
	repo *repository.OrderRepo
}

func NewOrderService() *OrderService {
	return &OrderService{repo: repository.NewOrderRepo()}
}

func (s *OrderService) List(ctx context.Context, page, pageSize int, filters map[string]string) ([]model.Order, int, error) {
	return s.repo.List(ctx, page, pageSize, filters)
}

func (s *OrderService) GetByID(ctx context.Context, id int64) (*model.Order, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *OrderService) GetStats(ctx context.Context, startDate, endDate string) (*model.OrderStats, error) {
	return s.repo.GetStats(ctx, startDate, endDate)
}

func (s *OrderService) GetRevenue(ctx context.Context, startDate, endDate, groupBy string) ([]model.RevenueRecord, error) {
	return s.repo.GetRevenueByDate(ctx, startDate, endDate, groupBy)
}
