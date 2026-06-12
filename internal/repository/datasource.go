package repository

import (
	"context"
	"time"

	"github.com/CodingFervor/live-commerce-bi/internal/database"
	"github.com/CodingFervor/live-commerce-bi/internal/model"
)

type DataSourceRepo struct{}

func NewDataSourceRepo() *DataSourceRepo { return &DataSourceRepo{} }

func (r *DataSourceRepo) Create(ctx context.Context, ds *model.DataSource) error {
	query := `INSERT INTO data_sources (name, platform, config, sync_interval, created_by)
		VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at`
	return database.Get().QueryRow(ctx, query,
		ds.Name, ds.Platform, ds.Config, ds.SyncInterval, ds.CreatedBy,
	).Scan(&ds.ID, &ds.CreatedAt)
}

func (r *DataSourceRepo) List(ctx context.Context, page, pageSize int) ([]model.DataSource, int, error) {
	var total int
	database.Get().QueryRow(ctx, "SELECT COUNT(*) FROM data_sources").Scan(&total)

	offset := (page - 1) * pageSize
	rows, err := database.Get().Query(ctx,
		`SELECT id, name, platform, config, status, last_sync_at, sync_interval, created_by, created_at, updated_at
		FROM data_sources ORDER BY created_at DESC LIMIT $1 OFFSET $2`, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []model.DataSource
	for rows.Next() {
		var ds model.DataSource
		if err := rows.Scan(&ds.ID, &ds.Name, &ds.Platform, &ds.Config, &ds.Status, &ds.LastSyncAt, &ds.SyncInterval, &ds.CreatedBy, &ds.CreatedAt, &ds.UpdatedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, ds)
	}
	return list, total, nil
}

func (r *DataSourceRepo) GetByID(ctx context.Context, id int64) (*model.DataSource, error) {
	var ds model.DataSource
	query := `SELECT id, name, platform, config, status, last_sync_at, sync_interval, created_by, created_at, updated_at
		FROM data_sources WHERE id = $1`
	err := database.Get().QueryRow(ctx, query, id).Scan(
		&ds.ID, &ds.Name, &ds.Platform, &ds.Config, &ds.Status, &ds.LastSyncAt, &ds.SyncInterval, &ds.CreatedBy, &ds.CreatedAt, &ds.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &ds, nil
}

func (r *DataSourceRepo) Update(ctx context.Context, ds *model.DataSource) error {
	_, err := database.Get().Exec(ctx,
		`UPDATE data_sources SET name=COALESCE($1,name), config=COALESCE($2,config),
		sync_interval=COALESCE($3,sync_interval), status=COALESCE($4,status), updated_at=$5 WHERE id=$6`,
		ds.Name, ds.Config, ds.SyncInterval, ds.Status, time.Now(), ds.ID)
	return err
}

func (r *DataSourceRepo) Delete(ctx context.Context, id int64) error {
	_, err := database.Get().Exec(ctx, "DELETE FROM data_sources WHERE id = $1", id)
	return err
}

func (r *DataSourceRepo) UpdateSyncStatus(ctx context.Context, id int64, status string) error {
	_, err := database.Get().Exec(ctx,
		"UPDATE data_sources SET last_sync_at = $1, status = $2 WHERE id = $3",
		time.Now(), status, id)
	return err
}
