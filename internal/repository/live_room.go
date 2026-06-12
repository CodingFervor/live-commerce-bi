package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/CodingFervor/live-commerce-bi/internal/database"
	"github.com/CodingFervor/live-commerce-bi/internal/model"
)

type LiveRoomRepo struct{}

func NewLiveRoomRepo() *LiveRoomRepo { return &LiveRoomRepo{} }

func (r *LiveRoomRepo) Create(ctx context.Context, lr *model.LiveRoom) error {
	query := `INSERT INTO live_rooms (streamer_id, platform, room_id, title, tags, data_source_id)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at`
	return database.Get().QueryRow(ctx, query,
		lr.StreamerID, lr.Platform, lr.RoomID, lr.Title,
		lr.Tags, lr.DataSourceID,
	).Scan(&lr.ID, &lr.CreatedAt)
}

func (r *LiveRoomRepo) List(ctx context.Context, page, pageSize int, filters map[string]string) ([]model.LiveRoom, int, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	if v, ok := filters["platform"]; ok && v != "" {
		conditions = append(conditions, fmt.Sprintf("lr.platform = $%d", argIdx))
		args = append(args, v)
		argIdx++
	}
	if v, ok := filters["status"]; ok && v != "" {
		conditions = append(conditions, fmt.Sprintf("lr.status = $%d", argIdx))
		args = append(args, v)
		argIdx++
	}
	if v, ok := filters["streamer_id"]; ok && v != "" {
		conditions = append(conditions, fmt.Sprintf("lr.streamer_id = $%d", argIdx))
		args = append(args, v)
		argIdx++
	}
	if v, ok := filters["start_date"]; ok && v != "" {
		conditions = append(conditions, fmt.Sprintf("lr.started_at >= $%d", argIdx))
		args = append(args, v)
		argIdx++
	}
	if v, ok := filters["end_date"]; ok && v != "" {
		conditions = append(conditions, fmt.Sprintf("lr.started_at <= $%d", argIdx))
		args = append(args, v+" 23:59:59")
		argIdx++
	}
	if v, ok := filters["search"]; ok && v != "" {
		conditions = append(conditions, fmt.Sprintf("(lr.title ILIKE $%d OR s.name ILIKE $%d)", argIdx, argIdx))
		args = append(args, "%"+v+"%")
		argIdx++
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	var total int
	countQ := "SELECT COUNT(*) FROM live_rooms lr JOIN streamers s ON lr.streamer_id = s.id " + where
	database.Get().QueryRow(ctx, countQ, args...).Scan(&total)

	offset := (page - 1) * pageSize
	query := `SELECT lr.id, lr.streamer_id, lr.platform, lr.room_id, lr.title, lr.status,
		lr.started_at, lr.ended_at, lr.duration, lr.peak_viewers, lr.avg_viewers,
		lr.total_views, lr.total_likes, lr.total_comments, lr.total_shares,
		lr.gmv, lr.order_count, lr.product_count, lr.conversion_rate,
		lr.data_source_id, lr.created_at, lr.updated_at, s.name AS streamer_name
		FROM live_rooms lr JOIN streamers s ON lr.streamer_id = s.id ` + where +
		` ORDER BY lr.created_at DESC LIMIT $` + fmt.Sprintf("%d", argIdx) +
		` OFFSET $` + fmt.Sprintf("%d", argIdx+1)
	args = append(args, pageSize, offset)

	rows, err := database.Get().Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []model.LiveRoom
	for rows.Next() {
		var lr model.LiveRoom
		if err := rows.Scan(&lr.ID, &lr.StreamerID, &lr.Platform, &lr.RoomID, &lr.Title, &lr.Status,
			&lr.StartedAt, &lr.EndedAt, &lr.Duration, &lr.PeakViewers, &lr.AvgViewers,
			&lr.TotalViews, &lr.TotalLikes, &lr.TotalComments, &lr.TotalShares,
			&lr.GMV, &lr.OrderCount, &lr.ProductCount, &lr.ConversionRate,
			&lr.DataSourceID, &lr.CreatedAt, &lr.UpdatedAt, &lr.StreamerName); err != nil {
			return nil, 0, err
		}
		list = append(list, lr)
	}
	return list, total, nil
}

func (r *LiveRoomRepo) GetByID(ctx context.Context, id int64) (*model.LiveRoom, error) {
	var lr model.LiveRoom
	query := `SELECT lr.id, lr.streamer_id, lr.platform, lr.room_id, lr.title, lr.status,
		lr.started_at, lr.ended_at, lr.duration, lr.peak_viewers, lr.avg_viewers,
		lr.total_views, lr.total_likes, lr.total_comments, lr.total_shares,
		lr.gmv, lr.order_count, lr.product_count, lr.conversion_rate,
		lr.tags, lr.data_source_id, lr.created_at, lr.updated_at, s.name AS streamer_name
		FROM live_rooms lr JOIN streamers s ON lr.streamer_id = s.id WHERE lr.id = $1`
	err := database.Get().QueryRow(ctx, query, id).Scan(
		&lr.ID, &lr.StreamerID, &lr.Platform, &lr.RoomID, &lr.Title, &lr.Status,
		&lr.StartedAt, &lr.EndedAt, &lr.Duration, &lr.PeakViewers, &lr.AvgViewers,
		&lr.TotalViews, &lr.TotalLikes, &lr.TotalComments, &lr.TotalShares,
		&lr.GMV, &lr.OrderCount, &lr.ProductCount, &lr.ConversionRate,
		&lr.Tags, &lr.DataSourceID, &lr.CreatedAt, &lr.UpdatedAt, &lr.StreamerName,
	)
	if err != nil {
		return nil, err
	}
	return &lr, nil
}

func (r *LiveRoomRepo) UpdateMetrics(ctx context.Context, id int64, metrics map[string]interface{}) error {
	setClauses := []string{"updated_at = $1"}
	args := []interface{}{time.Now()}
	i := 2
	for k, v := range metrics {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", k, i))
		args = append(args, v)
		i++
	}
	args = append(args, id)
	query := "UPDATE live_rooms SET " + strings.Join(setClauses, ", ") + fmt.Sprintf(" WHERE id = $%d", i)
	_, err := database.Get().Exec(ctx, query, args...)
	return err
}

func (r *LiveRoomRepo) GetActiveRooms(ctx context.Context) ([]model.LiveRoom, error) {
	rows, err := database.Get().Query(ctx,
		`SELECT lr.id, lr.streamer_id, lr.platform, lr.room_id, lr.title, lr.status,
		lr.started_at, lr.ended_at, lr.duration, lr.peak_viewers, lr.avg_viewers,
		lr.total_views, lr.total_likes, lr.total_comments, lr.total_shares,
		lr.gmv, lr.order_count, lr.product_count, lr.conversion_rate,
		lr.data_source_id, lr.created_at, lr.updated_at, s.name AS streamer_name
		FROM live_rooms lr JOIN streamers s ON lr.streamer_id = s.id
		WHERE lr.status = 'live' ORDER BY lr.gmv DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []model.LiveRoom
	for rows.Next() {
		var lr model.LiveRoom
		if err := rows.Scan(&lr.ID, &lr.StreamerID, &lr.Platform, &lr.RoomID, &lr.Title, &lr.Status,
			&lr.StartedAt, &lr.EndedAt, &lr.Duration, &lr.PeakViewers, &lr.AvgViewers,
			&lr.TotalViews, &lr.TotalLikes, &lr.TotalComments, &lr.TotalShares,
			&lr.GMV, &lr.OrderCount, &lr.ProductCount, &lr.ConversionRate,
			&lr.DataSourceID, &lr.CreatedAt, &lr.UpdatedAt, &lr.StreamerName); err != nil {
			return nil, err
		}
		list = append(list, lr)
	}
	return list, nil
}

func (r *LiveRoomRepo) GetTopByGMV(ctx context.Context, limit int, startDate, endDate string) ([]model.LiveRoom, error) {
	args := []interface{}{limit}
	query := `SELECT lr.id, lr.streamer_id, lr.platform, lr.room_id, lr.title, lr.status,
		lr.started_at, lr.duration, lr.peak_viewers, lr.avg_viewers,
		lr.total_views, lr.gmv, lr.order_count, lr.conversion_rate,
		lr.created_at, s.name AS streamer_name
		FROM live_rooms lr JOIN streamers s ON lr.streamer_id = s.id
		WHERE lr.status = 'ended'`
	if startDate != "" {
		query += fmt.Sprintf(" AND lr.started_at >= $%d", len(args)+1)
		args = append(args, startDate)
	}
	if endDate != "" {
		query += fmt.Sprintf(" AND lr.started_at <= $%d", len(args)+1)
		args = append(args, endDate+" 23:59:59")
	}
	query += " ORDER BY lr.gmv DESC LIMIT $1"

	rows, err := database.Get().Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []model.LiveRoom
	for rows.Next() {
		var lr model.LiveRoom
		if err := rows.Scan(&lr.ID, &lr.StreamerID, &lr.Platform, &lr.RoomID, &lr.Title, &lr.Status,
			&lr.StartedAt, &lr.Duration, &lr.PeakViewers, &lr.AvgViewers,
			&lr.TotalViews, &lr.GMV, &lr.OrderCount, &lr.ConversionRate,
			&lr.CreatedAt, &lr.StreamerName); err != nil {
			return nil, err
		}
		list = append(list, lr)
	}
	return list, nil
}
