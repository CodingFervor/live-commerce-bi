package service

import (
	"context"
	"encoding/json"

	"github.com/CodingFervor/live-commerce-bi/internal/cache"
	"github.com/CodingFervor/live-commerce-bi/internal/model"
	"github.com/CodingFervor/live-commerce-bi/internal/repository"
	"time"
)

type LiveRoomService struct {
	repo        *repository.LiveRoomRepo
	analyticsRepo *repository.AnalyticsRepo
}

func NewLiveRoomService() *LiveRoomService {
	return &LiveRoomService{
		repo:          repository.NewLiveRoomRepo(),
		analyticsRepo: repository.NewAnalyticsRepo(),
	}
}

func (s *LiveRoomService) Create(ctx context.Context, req *model.LiveRoomCreate) (*model.LiveRoom, error) {
	lr := &model.LiveRoom{
		StreamerID:   req.StreamerID,
		Platform:     req.Platform,
		RoomID:       req.RoomID,
		Title:        req.Title,
		Tags:         req.Tags,
		DataSourceID: req.DataSourceID,
		Status:       "scheduled",
	}
	if err := s.repo.Create(ctx, lr); err != nil {
		return nil, err
	}
	return lr, nil
}

func (s *LiveRoomService) List(ctx context.Context, page, pageSize int, filters map[string]string) ([]model.LiveRoom, int, error) {
	return s.repo.List(ctx, page, pageSize, filters)
}

func (s *LiveRoomService) GetByID(ctx context.Context, id int64) (*model.LiveRoom, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *LiveRoomService) GetMetrics(ctx context.Context, id int64) (*model.LiveRoomMetrics, error) {
	lr, err := s.repo.GetByID(ctx, id)
	if err != nil { return nil, err }

	// Try Redis cache first
	rdb := cache.Get()
	cached, err := rdb.HGetAll(ctx, "live:metrics:"+string(rune(id))).Result()
	_ = cached

	metrics := &model.LiveRoomMetrics{
		RoomID:         lr.ID,
		PeakViewers:    lr.PeakViewers,
		TotalViews:     lr.TotalViews,
		TotalLikes:     lr.TotalLikes,
		TotalComments:  lr.TotalComments,
		GMV:            lr.GMV,
		OrderCount:     lr.OrderCount,
		ConversionRate: lr.ConversionRate,
	}

	// Get latest viewer metrics for concurrent
	vmList, _ := s.analyticsRepo.GetViewerMetrics(ctx, id, 1)
	if len(vmList) > 0 {
		metrics.ConcurrentIndex = vmList[0].ConcurrentIndex
		metrics.AvgWatchTime = vmList[0].AvgWatchTime
	}

	return metrics, nil
}

func (s *LiveRoomService) GetActiveRooms(ctx context.Context) ([]model.LiveRoom, error) {
	return s.repo.GetActiveRooms(ctx)
}

func (s *LiveRoomService) GetProducts(ctx context.Context, roomID int64) ([]interface{}, error) {
	// Query live_room_products joined with products
	rows, err := cache.Get().Get(ctx, "live:products:"+string(rune(roomID))).Result()
	if err == nil {
		var result []interface{}
		json.Unmarshal([]byte(rows), &result)
		return result, nil
	}
	// Fallback: return basic data
	return []interface{}{}, nil
}

func (s *LiveRoomService) GetFunnel(ctx context.Context, roomID int64) ([]model.ConversionAnalysis, error) {
	return s.analyticsRepo.GetConversionFunnel(ctx, roomID)
}

func (s *LiveRoomService) UpdateRealtimeMetrics(ctx context.Context, id int64) error {
	// Simulate real-time metrics update
	rdb := cache.Get()
	now := time.Now().Unix()
	data := map[string]interface{}{
		"room_id":    id,
		"updated_at": now,
	}
	b, _ := json.Marshal(data)
	return rdb.Publish(ctx, "live:metrics:update", string(b)).Err()
}
