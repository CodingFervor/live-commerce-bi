package service

import (
	"context"

	"github.com/CodingFervor/live-commerce-bi/internal/model"
	"github.com/CodingFervor/live-commerce-bi/internal/repository"
)

type DashboardService struct {
	repo *repository.DashboardRepo
}

func NewDashboardService() *DashboardService {
	return &DashboardService{repo: repository.NewDashboardRepo()}
}

func (s *DashboardService) Create(ctx context.Context, req *model.DashboardCreate, userID int64) (*model.Dashboard, error) {
	d := &model.Dashboard{
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		Layout:      req.Layout,
		IsDefault:   req.IsDefault,
		OwnerID:     userID,
	}
	if d.Type == "" { d.Type = "custom" }
	if d.Layout == "" { d.Layout = "{}" }
	if err := s.repo.Create(ctx, d); err != nil {
		return nil, err
	}
	return d, nil
}

func (s *DashboardService) List(ctx context.Context, userID int64) ([]model.Dashboard, error) {
	return s.repo.List(ctx, userID)
}

func (s *DashboardService) GetByID(ctx context.Context, id int64) (*model.Dashboard, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *DashboardService) Update(ctx context.Context, id int64, req *model.DashboardCreate) (*model.Dashboard, error) {
	d, err := s.repo.GetByID(ctx, id)
	if err != nil { return nil, err }
	if req.Name != "" { d.Name = req.Name }
	if req.Description != "" { d.Description = req.Description }
	if req.Layout != "" { d.Layout = req.Layout }
	if err := s.repo.Update(ctx, d); err != nil {
		return nil, err
	}
	return d, nil
}

func (s *DashboardService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func (s *DashboardService) CreateWidget(ctx context.Context, dashboardID int64, req *model.WidgetCreate) (*model.DashboardWidget, error) {
	w := &model.DashboardWidget{
		DashboardID:     dashboardID,
		Title:           req.Title,
		Type:            req.Type,
		DataSource:      req.DataSource,
		Config:          req.Config,
		Position:        req.Position,
		RefreshInterval: req.RefreshInterval,
	}
	if w.Config == "" { w.Config = "{}" }
	if w.Position == "" { w.Position = "{}" }
	if w.RefreshInterval == 0 { w.RefreshInterval = 30 }
	if err := s.repo.CreateWidget(ctx, w); err != nil {
		return nil, err
	}
	return w, nil
}

func (s *DashboardService) ListWidgets(ctx context.Context, dashboardID int64) ([]model.DashboardWidget, error) {
	return s.repo.ListWidgets(ctx, dashboardID)
}

func (s *DashboardService) UpdateWidget(ctx context.Context, id int64, req *model.WidgetCreate) (*model.DashboardWidget, error) {
	// simplified: create new with updated fields
	w := &model.DashboardWidget{
		ID:              id,
		Title:           req.Title,
		Type:            req.Type,
		DataSource:      req.DataSource,
		Config:          req.Config,
		Position:        req.Position,
		RefreshInterval: req.RefreshInterval,
	}
	if err := s.repo.UpdateWidget(ctx, w); err != nil {
		return nil, err
	}
	return w, nil
}

func (s *DashboardService) DeleteWidget(ctx context.Context, id int64) error {
	return s.repo.DeleteWidget(ctx, id)
}
