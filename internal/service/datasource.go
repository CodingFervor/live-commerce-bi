package service

import (
	"context"
	"fmt"

	"github.com/CodingFervor/live-commerce-bi/internal/cache"
	"github.com/CodingFervor/live-commerce-bi/internal/model"
	"github.com/CodingFervor/live-commerce-bi/internal/repository"
)

type DataSourceService struct {
	repo     *repository.DataSourceRepo
	syncRepo *repository.SyncLogRepo
}

func NewDataSourceService() *DataSourceService {
	return &DataSourceService{repo: repository.NewDataSourceRepo()}
}

func (s *DataSourceService) Create(ctx context.Context, req *model.DataSourceCreate, userID int64) (*model.DataSource, error) {
	ds := &model.DataSource{
		Name:         req.Name,
		Platform:     req.Platform,
		Config:       req.Config,
		SyncInterval: req.SyncInterval,
		CreatedBy:    userID,
	}
	if ds.SyncInterval == 0 { ds.SyncInterval = 300 }
	if err := s.repo.Create(ctx, ds); err != nil {
		return nil, err
	}
	return ds, nil
}

func (s *DataSourceService) List(ctx context.Context, page, pageSize int) ([]model.DataSource, int, error) {
	return s.repo.List(ctx, page, pageSize)
}

func (s *DataSourceService) GetByID(ctx context.Context, id int64) (*model.DataSource, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *DataSourceService) Update(ctx context.Context, id int64, req *model.DataSourceUpdate) (*model.DataSource, error) {
	ds, err := s.repo.GetByID(ctx, id)
	if err != nil { return nil, err }
	if req.Name != nil { ds.Name = *req.Name }
	if req.Config != nil { ds.Config = *req.Config }
	if req.SyncInterval != nil { ds.SyncInterval = *req.SyncInterval }
	if req.Status != nil { ds.Status = *req.Status }
	if err := s.repo.Update(ctx, ds); err != nil {
		return nil, err
	}
	return ds, nil
}

func (s *DataSourceService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func (s *DataSourceService) TriggerSync(ctx context.Context, id int64) error {
	ds, err := s.repo.GetByID(ctx, id)
	if err != nil { return err }
	// In production: dispatch sync job to worker queue
	// For now: update status and cache
	_ = s.repo.UpdateSyncStatus(ctx, id, "active")
	cache.Get().Publish(ctx, "sync:trigger", fmt.Sprintf(`{"id":%d,"platform":"%s"}`, id, ds.Platform))
	return nil
}

func (s *DataSourceService) GetStatus(ctx context.Context, id int64) (map[string]interface{}, error) {
	ds, err := s.repo.GetByID(ctx, id)
	if err != nil { return nil, err }
	return map[string]interface{}{
		"id":            ds.ID,
		"name":          ds.Name,
		"platform":      ds.Platform,
		"status":        ds.Status,
		"last_sync_at":  ds.LastSyncAt,
		"sync_interval": ds.SyncInterval,
	}, nil
}
