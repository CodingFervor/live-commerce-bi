package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/CodingFervor/live-commerce-bi/internal/database"
	"github.com/CodingFervor/live-commerce-bi/internal/model"
)

type StreamerRepo struct{}

func NewStreamerRepo() *StreamerRepo { return &StreamerRepo{} }

func (r *StreamerRepo) Create(ctx context.Context, s *model.Streamer) error {
	query := `INSERT INTO streamers (name, platform, platform_id, avatar, category, tags, data_source_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, created_at`
	return database.Get().QueryRow(ctx, query,
		s.Name, s.Platform, s.PlatformID, s.Avatar, s.Category, s.Tags, s.DataSourceID,
	).Scan(&s.ID, &s.CreatedAt)
}

func (r *StreamerRepo) List(ctx context.Context, page, pageSize int, platform, category string) ([]model.Streamer, int, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	if platform != "" {
		conditions = append(conditions, fmt.Sprintf("platform = $%d", argIdx))
		args = append(args, platform)
		argIdx++
	}
	if category != "" {
		conditions = append(conditions, fmt.Sprintf("category = $%d", argIdx))
		args = append(args, category)
		argIdx++
	}
	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	var total int
	database.Get().QueryRow(ctx, "SELECT COUNT(*) FROM streamers "+where, args...).Scan(&total)

	offset := (page - 1) * pageSize
	query := `SELECT id, name, platform, platform_id, avatar, follower_count, category, tags, status, data_source_id, created_at, updated_at
		FROM streamers ` + where + ` ORDER BY follower_count DESC LIMIT $` + fmt.Sprintf("%d", argIdx) + ` OFFSET $` + fmt.Sprintf("%d", argIdx+1)
	args = append(args, pageSize, offset)

	rows, err := database.Get().Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []model.Streamer
	for rows.Next() {
		var s model.Streamer
		if err := rows.Scan(&s.ID, &s.Name, &s.Platform, &s.PlatformID, &s.Avatar, &s.FollowerCount, &s.Category, &s.Tags, &s.Status, &s.DataSourceID, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, s)
	}
	return list, total, nil
}

func (r *StreamerRepo) GetByID(ctx context.Context, id int64) (*model.Streamer, error) {
	var s model.Streamer
	err := database.Get().QueryRow(ctx,
		`SELECT id, name, platform, platform_id, avatar, follower_count, category, tags, status, data_source_id, created_at, updated_at
		FROM streamers WHERE id = $1`, id).Scan(
		&s.ID, &s.Name, &s.Platform, &s.PlatformID, &s.Avatar, &s.FollowerCount, &s.Category, &s.Tags, &s.Status, &s.DataSourceID, &s.CreatedAt, &s.UpdatedAt)
	return &s, err
}

func (r *StreamerRepo) Update(ctx context.Context, s *model.Streamer) error {
	_, err := database.Get().Exec(ctx,
		`UPDATE streamers SET name=$1, avatar=$2, category=$3, follower_count=$4 WHERE id=$5`,
		s.Name, s.Avatar, s.Category, s.FollowerCount, s.ID)
	return err
}

func (r *StreamerRepo) GetRankings(ctx context.Context, metric string, limit int, platform string) ([]model.StreamerRanking, error) {
	orderCol := "total_gmv"
	switch metric {
	case "orders":
		orderCol = "total_orders"
	case "views":
		orderCol = "total_views"
	case "conversion":
		orderCol = "avg_conversion"
	}

	args := []interface{}{limit}
	argIdx := 1
	where := ""
	if platform != "" {
		where = fmt.Sprintf("WHERE s.platform = $%d", argIdx+1)
		args = append(args, platform)
	}

	query := fmt.Sprintf(`SELECT s.id, s.name, s.platform, s.avatar, s.category,
		COALESCE(SUM(lr.gmv),0) AS total_gmv,
		COALESCE(SUM(lr.order_count),0) AS total_orders,
		COALESCE(SUM(lr.total_views),0) AS total_views,
		COALESCE(AVG(lr.conversion_rate),0) AS avg_conversion
		FROM streamers s LEFT JOIN live_rooms lr ON s.id = lr.streamer_id
		%s GROUP BY s.id ORDER BY %s DESC LIMIT $1`, where, orderCol)

	rows, err := database.Get().Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []model.StreamerRanking
	rank := 1
	for rows.Next() {
		var sr model.StreamerRanking
		if err := rows.Scan(&sr.StreamerID, &sr.StreamerName, &sr.Platform, &sr.Avatar, &sr.Category,
			&sr.GMV, &sr.OrderCount, &sr.ViewerCount, &sr.ConversionRate); err != nil {
			return nil, err
		}
		sr.Rank = rank
		list = append(list, sr)
		rank++
	}
	return list, nil
}

func (r *StreamerRepo) GetPerformance(ctx context.Context, id int64) (*model.StreamerPerformance, error) {
	var perf model.StreamerPerformance
	s, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	perf.Streamer = *s
	err = database.Get().QueryRow(ctx,
		`SELECT COUNT(DISTINCT id), COALESCE(SUM(gmv),0), COALESCE(SUM(order_count),0),
		COALESCE(SUM(total_views),0), COALESCE(AVG(conversion_rate),0),
		COALESCE(AVG(avg_viewers),0), COALESCE(SUM(duration),0)
		FROM live_rooms WHERE streamer_id = $1`, id).Scan(
		&perf.TotalLiveRooms, &perf.TotalGMV, &perf.TotalOrders,
		&perf.TotalViews, &perf.AvgConversion, &perf.AvgViewers, &perf.TotalDuration)
	return &perf, err
}
