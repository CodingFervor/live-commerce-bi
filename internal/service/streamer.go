package service

import (
	"context"

	"github.com/CodingFervor/live-commerce-bi/internal/model"
	"github.com/CodingFervor/live-commerce-bi/internal/repository"
)

type StreamerService struct {
	repo *repository.StreamerRepo
}

func NewStreamerService() *StreamerService {
	return &StreamerService{repo: repository.NewStreamerRepo()}
}

func (s *StreamerService) Create(ctx context.Context, req *model.StreamerCreate) (*model.Streamer, error) {
	st := &model.Streamer{
		Name:         req.Name,
		Platform:     req.Platform,
		PlatformID:   req.PlatformID,
		Avatar:       req.Avatar,
		Category:     req.Category,
		Tags:         req.Tags,
		DataSourceID: req.DataSourceID,
	}
	if err := s.repo.Create(ctx, st); err != nil {
		return nil, err
	}
	return st, nil
}

func (s *StreamerService) List(ctx context.Context, page, pageSize int, platform, category string) ([]model.Streamer, int, error) {
	return s.repo.List(ctx, page, pageSize, platform, category)
}

func (s *StreamerService) GetByID(ctx context.Context, id int64) (*model.Streamer, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *StreamerService) Update(ctx context.Context, id int64, req *model.StreamerCreate) (*model.Streamer, error) {
	st, err := s.repo.GetByID(ctx, id)
	if err != nil { return nil, err }
	if req.Name != "" { st.Name = req.Name }
	if req.Avatar != "" { st.Avatar = req.Avatar }
	if req.Category != "" { st.Category = req.Category }
	if err := s.repo.Update(ctx, st); err != nil {
		return nil, err
	}
	return st, nil
}

func (s *StreamerService) GetRankings(ctx context.Context, metric string, limit int, platform string) ([]model.StreamerRanking, error) {
	if limit <= 0 { limit = 20 }
	return s.repo.GetRankings(ctx, metric, limit, platform)
}

func (s *StreamerService) GetPerformance(ctx context.Context, id int64) (*model.StreamerPerformance, error) {
	return s.repo.GetPerformance(ctx, id)
}
