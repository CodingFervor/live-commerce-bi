package repository

import (
	"context"

	"github.com/CodingFervor/live-commerce-bi/internal/database"
	"github.com/CodingFervor/live-commerce-bi/internal/model"
)

type ReportRepo struct{}

func NewReportRepo() *ReportRepo { return &ReportRepo{} }

func (r *ReportRepo) CreateTemplate(ctx context.Context, t *model.ReportTemplate) error {
	return database.Get().QueryRow(ctx,
		`INSERT INTO report_templates (name, description, type, sections, thumbnail, is_public, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id, created_at`,
		t.Name, t.Description, t.Type, t.Sections, t.Thumbnail, t.IsPublic, t.CreatedBy,
	).Scan(&t.ID, &t.CreatedAt)
}

func (r *ReportRepo) ListTemplates(ctx context.Context) ([]model.ReportTemplate, error) {
	rows, err := database.Get().Query(ctx,
		`SELECT id, name, description, type, sections, thumbnail, is_public, created_by, created_at, updated_at
		FROM report_templates WHERE is_public=true ORDER BY created_at DESC`)
	if err != nil { return nil, err }
	defer rows.Close()
	var list []model.ReportTemplate
	for rows.Next() {
		var t model.ReportTemplate
		if err := rows.Scan(&t.ID, &t.Name, &t.Description, &t.Type, &t.Sections, &t.Thumbnail, &t.IsPublic, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt); err != nil {
			continue
		}
		list = append(list, t)
	}
	return list, nil
}

func (r *ReportRepo) CreateReport(ctx context.Context, rp *model.Report) error {
	return database.Get().QueryRow(ctx,
		`INSERT INTO reports (name, template_id, type, status, config, schedule, date_range_start, date_range_end, filters, output_format, generated_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING id, created_at`,
		rp.Name, rp.TemplateID, rp.Type, rp.Status, rp.Config, rp.Schedule,
		rp.DateRangeStart, rp.DateRangeEnd, rp.Filters, rp.OutputFormat, rp.GeneratedBy,
	).Scan(&rp.ID, &rp.CreatedAt)
}

func (r *ReportRepo) ListReports(ctx context.Context, page, pageSize int) ([]model.Report, int, error) {
	var total int
	database.Get().QueryRow(ctx, "SELECT COUNT(*) FROM reports").Scan(&total)
	offset := (page - 1) * pageSize
	rows, err := database.Get().Query(ctx,
		`SELECT id, name, template_id, type, status, config, schedule, date_range_start, date_range_end,
		filters, output_format, file_path, generated_at, generated_by, created_at, updated_at
		FROM reports ORDER BY created_at DESC LIMIT $1 OFFSET $2`, pageSize, offset)
	if err != nil { return nil, 0, err }
	defer rows.Close()
	var list []model.Report
	for rows.Next() {
		var rp model.Report
		if err := rows.Scan(&rp.ID, &rp.Name, &rp.TemplateID, &rp.Type, &rp.Status, &rp.Config, &rp.Schedule,
			&rp.DateRangeStart, &rp.DateRangeEnd, &rp.Filters, &rp.OutputFormat, &rp.FilePath, &rp.GeneratedAt,
			&rp.GeneratedBy, &rp.CreatedAt, &rp.UpdatedAt); err != nil {
			continue
		}
		list = append(list, rp)
	}
	return list, total, nil
}

func (r *ReportRepo) GetReport(ctx context.Context, id int64) (*model.Report, error) {
	var rp model.Report
	err := database.Get().QueryRow(ctx,
		`SELECT id, name, template_id, type, status, config, schedule, date_range_start, date_range_end,
		filters, output_format, file_path, generated_at, generated_by, created_at, updated_at
		FROM reports WHERE id=$1`, id).Scan(
		&rp.ID, &rp.Name, &rp.TemplateID, &rp.Type, &rp.Status, &rp.Config, &rp.Schedule,
		&rp.DateRangeStart, &rp.DateRangeEnd, &rp.Filters, &rp.OutputFormat, &rp.FilePath, &rp.GeneratedAt,
		&rp.GeneratedBy, &rp.CreatedAt, &rp.UpdatedAt)
	return &rp, err
}

func (r *ReportRepo) UpdateReport(ctx context.Context, rp *model.Report) error {
	_, err := database.Get().Exec(ctx,
		`UPDATE reports SET name=$1, status=$2, config=$3, schedule=$4, filters=$5 WHERE id=$6`,
		rp.Name, rp.Status, rp.Config, rp.Schedule, rp.Filters, rp.ID)
	return err
}

func (r *ReportRepo) DeleteReport(ctx context.Context, id int64) error {
	_, err := database.Get().Exec(ctx, "DELETE FROM reports WHERE id=$1", id)
	return err
}
