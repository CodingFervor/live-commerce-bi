package repository

import (
	"context"
	"time"

	"github.com/CodingFervor/live-commerce-bi/internal/database"
	"github.com/CodingFervor/live-commerce-bi/internal/model"
)

type CompassRepo struct{}

func NewCompassRepo() *CompassRepo { return &CompassRepo{} }

// ─── Session CRUD ───

func (r *CompassRepo) CreateSession(ctx context.Context, s *model.CompassSession) error {
	return database.Get().QueryRow(ctx,
		`INSERT INTO compass_sessions (name, cookie, shop_id, shop_name, user_agent, proxy_url, status, max_daily_reqs, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,'active',$7,$8) RETURNING id, created_at`,
		s.Name, s.Cookie, s.ShopID, s.ShopName, s.UserAgent, s.ProxyURL, s.MaxDailyReqs, s.CreatedBy,
	).Scan(&s.ID, &s.CreatedAt)
}

func (r *CompassRepo) GetSession(ctx context.Context, id int64) (*model.CompassSession, error) {
	s := &model.CompassSession{}
	err := database.Get().QueryRow(ctx,
		`SELECT id, name, cookie, shop_id, shop_name, user_agent, proxy_url, status,
		last_active_at, daily_requests, max_daily_reqs, fail_count, next_available_at, created_by, created_at, updated_at
		FROM compass_sessions WHERE id=$1`, id).Scan(
		&s.ID, &s.Name, &s.Cookie, &s.ShopID, &s.ShopName, &s.UserAgent, &s.ProxyURL, &s.Status,
		&s.LastActiveAt, &s.DailyRequests, &s.MaxDailyReqs, &s.FailCount, &s.NextAvailableAt,
		&s.CreatedBy, &s.CreatedAt, &s.UpdatedAt)
	return s, err
}

func (r *CompassRepo) ListSessions(ctx context.Context, page, pageSize int) ([]model.CompassSession, int, error) {
	var total int
	database.Get().QueryRow(ctx, "SELECT COUNT(*) FROM compass_sessions").Scan(&total)
	offset := (page - 1) * pageSize
	rows, err := database.Get().Query(ctx,
		`SELECT id, name, shop_id, shop_name, user_agent, proxy_url, status,
		last_active_at, daily_requests, max_daily_reqs, fail_count, next_available_at, created_by, created_at, updated_at
		FROM compass_sessions ORDER BY created_at DESC LIMIT $1 OFFSET $2`, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []model.CompassSession
	for rows.Next() {
		var s model.CompassSession
		if err := rows.Scan(&s.ID, &s.Name, &s.ShopID, &s.ShopName, &s.UserAgent, &s.ProxyURL, &s.Status,
			&s.LastActiveAt, &s.DailyRequests, &s.MaxDailyReqs, &s.FailCount, &s.NextAvailableAt,
			&s.CreatedBy, &s.CreatedAt, &s.UpdatedAt); err != nil {
			continue
		}
		list = append(list, s)
	}
	return list, total, nil
}

func (r *CompassRepo) UpdateSession(ctx context.Context, s *model.CompassSession) error {
	_, err := database.Get().Exec(ctx,
		`UPDATE compass_sessions SET name=$1, cookie=COALESCE($2,cookie), user_agent=$3,
		proxy_url=$4, status=$5, max_daily_reqs=$6 WHERE id=$7`,
		s.Name, s.Cookie, s.UserAgent, s.ProxyURL, s.Status, s.MaxDailyReqs, s.ID)
	return err
}

func (r *CompassRepo) DeleteSession(ctx context.Context, id int64) error {
	_, err := database.Get().Exec(ctx, "DELETE FROM compass_sessions WHERE id=$1", id)
	return err
}

func (r *CompassRepo) ListActiveSessions(ctx context.Context) ([]model.CompassSession, error) {
	rows, err := database.Get().Query(ctx,
		`SELECT id, name, cookie, shop_id, shop_name, user_agent, proxy_url, status,
		last_active_at, daily_requests, max_daily_reqs, fail_count, next_available_at, created_by, created_at, updated_at
		FROM compass_sessions WHERE status='active' ORDER BY last_active_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []model.CompassSession
	for rows.Next() {
		var s model.CompassSession
		if err := rows.Scan(&s.ID, &s.Name, &s.Cookie, &s.ShopID, &s.ShopName, &s.UserAgent, &s.ProxyURL, &s.Status,
			&s.LastActiveAt, &s.DailyRequests, &s.MaxDailyReqs, &s.FailCount, &s.NextAvailableAt,
			&s.CreatedBy, &s.CreatedAt, &s.UpdatedAt); err != nil {
			continue
		}
		list = append(list, s)
	}
	return list, nil
}

// ─── Task CRUD ───

func (r *CompassRepo) CreateTask(ctx context.Context, t *model.CompassTask) error {
	return database.Get().QueryRow(ctx,
		`INSERT INTO compass_tasks (session_id, task_type, params, status)
		VALUES ($1,$2,$3,'pending') RETURNING id, created_at`,
		t.SessionID, t.TaskType, t.Params,
	).Scan(&t.ID, &t.CreatedAt)
}

func (r *CompassRepo) GetTask(ctx context.Context, id int64) (*model.CompassTask, error) {
	t := &model.CompassTask{}
	err := database.Get().QueryRow(ctx,
		`SELECT id, session_id, task_type, params, status, result, records_count, error_msg, started_at, completed_at, created_at
		FROM compass_tasks WHERE id=$1`, id).Scan(
		&t.ID, &t.SessionID, &t.TaskType, &t.Params, &t.Status, &t.Result,
		&t.RecordsCount, &t.ErrorMsg, &t.StartedAt, &t.CompletedAt, &t.CreatedAt)
	return t, err
}

func (r *CompassRepo) ListTasks(ctx context.Context, sessionID int64, limit int) ([]model.CompassTask, error) {
	if limit <= 0 {
		limit = 20
	}
	args := []interface{}{limit}
	query := `SELECT id, session_id, task_type, params, status, result, records_count, error_msg, started_at, completed_at, created_at
		FROM compass_tasks`
	if sessionID > 0 {
		query += ` WHERE session_id=$2 ORDER BY created_at DESC LIMIT $1`
		args = append(args, sessionID)
	} else {
		query += ` ORDER BY created_at DESC LIMIT $1`
	}
	rows, err := database.Get().Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []model.CompassTask
	for rows.Next() {
		var t model.CompassTask
		if err := rows.Scan(&t.ID, &t.SessionID, &t.TaskType, &t.Params, &t.Status, &t.Result,
			&t.RecordsCount, &t.ErrorMsg, &t.StartedAt, &t.CompletedAt, &t.CreatedAt); err != nil {
			continue
		}
		list = append(list, t)
	}
	return list, nil
}

func (r *CompassRepo) UpdateTaskStatus(ctx context.Context, id int64, status, result, errMsg string, records int) error {
	now := time.Now()
	if status == "running" {
		_, err := database.Get().Exec(ctx,
			`UPDATE compass_tasks SET status=$1, started_at=$2 WHERE id=$3`,
			status, now, id)
		return err
	}
	_, err := database.Get().Exec(ctx,
		`UPDATE compass_tasks SET status=$1, result=$2, error_msg=$3, records_count=$4, completed_at=$5 WHERE id=$6`,
		status, result, errMsg, records, now, id)
	return err
}
