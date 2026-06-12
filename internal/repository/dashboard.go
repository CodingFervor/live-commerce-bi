package repository

import (
	"context"
	"encoding/json"

	"github.com/CodingFervor/live-commerce-bi/internal/database"
	"github.com/CodingFervor/live-commerce-bi/internal/model"
)

type DashboardRepo struct{}

func NewDashboardRepo() *DashboardRepo { return &DashboardRepo{} }

func (r *DashboardRepo) Create(ctx context.Context, d *model.Dashboard) error {
	return database.Get().QueryRow(ctx,
		`INSERT INTO dashboards (name, description, type, layout, is_default, owner_id)
		VALUES ($1,$2,$3,$4,$5,$6) RETURNING id, created_at`,
		d.Name, d.Description, d.Type, d.Layout, d.IsDefault, d.OwnerID,
	).Scan(&d.ID, &d.CreatedAt)
}

func (r *DashboardRepo) List(ctx context.Context, ownerID int64) ([]model.Dashboard, error) {
	rows, err := database.Get().Query(ctx,
		`SELECT id, name, description, type, layout, is_default, owner_id, status, created_at, updated_at
		FROM dashboards WHERE owner_id=$1 OR is_default=true ORDER BY is_default DESC, created_at DESC`, ownerID)
	if err != nil { return nil, err }
	defer rows.Close()
	var list []model.Dashboard
	for rows.Next() {
		var d model.Dashboard
		if err := rows.Scan(&d.ID, &d.Name, &d.Description, &d.Type, &d.Layout, &d.IsDefault, &d.OwnerID, &d.Status, &d.CreatedAt, &d.UpdatedAt); err != nil {
			continue
		}
		list = append(list, d)
	}
	return list, nil
}

func (r *DashboardRepo) GetByID(ctx context.Context, id int64) (*model.Dashboard, error) {
	var d model.Dashboard
	err := database.Get().QueryRow(ctx,
		`SELECT id, name, description, type, layout, is_default, owner_id, status, created_at, updated_at
		FROM dashboards WHERE id=$1`, id).Scan(
		&d.ID, &d.Name, &d.Description, &d.Type, &d.Layout, &d.IsDefault, &d.OwnerID, &d.Status, &d.CreatedAt, &d.UpdatedAt)
	return &d, err
}

func (r *DashboardRepo) Update(ctx context.Context, d *model.Dashboard) error {
	_, err := database.Get().Exec(ctx,
		`UPDATE dashboards SET name=$1, description=$2, layout=$3 WHERE id=$4`,
		d.Name, d.Description, d.Layout, d.ID)
	return err
}

func (r *DashboardRepo) Delete(ctx context.Context, id int64) error {
	_, err := database.Get().Exec(ctx, "DELETE FROM dashboards WHERE id=$1", id)
	return err
}

func (r *DashboardRepo) CreateWidget(ctx context.Context, w *model.DashboardWidget) error {
	return database.Get().QueryRow(ctx,
		`INSERT INTO dashboard_widgets (dashboard_id, title, type, data_source, config, position, refresh_interval)
		VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id, created_at`,
		w.DashboardID, w.Title, w.Type, w.DataSource, w.Config, w.Position, w.RefreshInterval,
	).Scan(&w.ID, &w.CreatedAt)
}

func (r *DashboardRepo) ListWidgets(ctx context.Context, dashboardID int64) ([]model.DashboardWidget, error) {
	rows, err := database.Get().Query(ctx,
		`SELECT id, dashboard_id, title, type, data_source, config, position, refresh_interval, created_at, updated_at
		FROM dashboard_widgets WHERE dashboard_id=$1 ORDER BY position`, dashboardID)
	if err != nil { return nil, err }
	defer rows.Close()
	var list []model.DashboardWidget
	for rows.Next() {
		var w model.DashboardWidget
		if err := rows.Scan(&w.ID, &w.DashboardID, &w.Title, &w.Type, &w.DataSource, &w.Config, &w.Position, &w.RefreshInterval, &w.CreatedAt, &w.UpdatedAt); err != nil {
			continue
		}
		list = append(list, w)
	}
	return list, nil
}

func (r *DashboardRepo) UpdateWidget(ctx context.Context, w *model.DashboardWidget) error {
	_, err := database.Get().Exec(ctx,
		`UPDATE dashboard_widgets SET title=$1, type=$2, config=$3, position=$4, refresh_interval=$5 WHERE id=$6`,
		w.Title, w.Type, w.Config, w.Position, w.RefreshInterval, w.ID)
	return err
}

func (r *DashboardRepo) DeleteWidget(ctx context.Context, id int64) error {
	_, err := database.Get().Exec(ctx, "DELETE FROM dashboard_widgets WHERE id=$1", id)
	return err
}

// helper
func toJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}
