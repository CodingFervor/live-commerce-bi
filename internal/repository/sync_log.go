package repository

import (
	"context"
	"time"

	"github.com/CodingFervor/live-commerce-bi/internal/database"
	"github.com/CodingFervor/live-commerce-bi/internal/model"
)

type SyncLogRepo struct{}

func NewSyncLogRepo() *SyncLogRepo { return &SyncLogRepo{} }

func (r *SyncLogRepo) Create(ctx context.Context, log *model.SyncLog) error {
	return database.Get().QueryRow(ctx,
		`INSERT INTO sync_logs (data_source_id, sync_type, status, started_at, records_processed, records_failed, error_message)
		VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id, created_at`,
		log.DataSourceID, log.SyncType, log.Status, log.StartedAt, log.RecordsProcessed, log.RecordsFailed, log.ErrorMessage,
	).Scan(&log.ID, &log.CreatedAt)
}

func (r *SyncLogRepo) ListBySource(ctx context.Context, sourceID int64, limit int) ([]model.SyncLog, error) {
	if limit <= 0 { limit = 20 }
	rows, err := database.Get().Query(ctx,
		`SELECT id, data_source_id, sync_type, status, started_at, completed_at,
		records_processed, records_failed, error_message, created_at
		FROM sync_logs WHERE data_source_id=$1 ORDER BY created_at DESC LIMIT $2`, sourceID, limit)
	if err != nil { return nil, err }
	defer rows.Close()
	var list []model.SyncLog
	for rows.Next() {
		var l model.SyncLog
		if err := rows.Scan(&l.ID, &l.DataSourceID, &l.SyncType, &l.Status, &l.StartedAt, &l.CompletedAt,
			&l.RecordsProcessed, &l.RecordsFailed, &l.ErrorMessage, &l.CreatedAt); err != nil {
			continue
		}
		list = append(list, l)
	}
	return list, nil
}

func (r *SyncLogRepo) Complete(ctx context.Context, id int64, processed, failed int, errMsg string) error {
	_, err := database.Get().Exec(ctx,
		"UPDATE sync_logs SET status='completed', completed_at=$1, records_processed=$2, records_failed=$3, error_message=$4 WHERE id=$5",
		time.Now(), processed, failed, errMsg, id)
	return err
}
