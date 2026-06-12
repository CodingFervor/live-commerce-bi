package service

import (
	"context"
	"fmt"
	"time"

	"github.com/CodingFervor/live-commerce-bi/internal/model"
	"github.com/CodingFervor/live-commerce-bi/internal/repository"
)

type ReportService struct {
	repo     *repository.ReportRepo
	tmplRepo *repository.ReportRepo
}

func NewReportService() *ReportService {
	return &ReportService{repo: repository.NewReportRepo()}
}

func (s *ReportService) CreateTemplate(ctx context.Context, t *model.ReportTemplate, userID int64) (*model.ReportTemplate, error) {
	t.CreatedBy = userID
	if t.Sections == "" { t.Sections = "[]" }
	if err := s.repo.CreateTemplate(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *ReportService) ListTemplates(ctx context.Context) ([]model.ReportTemplate, error) {
	return s.repo.ListTemplates(ctx)
}

func (s *ReportService) CreateReport(ctx context.Context, req *model.ReportCreate, userID int64) (*model.Report, error) {
	rp := &model.Report{
		Name:           req.Name,
		TemplateID:     req.TemplateID,
		Type:           req.Type,
		Status:         "draft",
		Config:         req.Config,
		Schedule:       req.Schedule,
		DateRangeStart: &req.DateRangeStart,
		DateRangeEnd:   &req.DateRangeEnd,
		Filters:        req.Filters,
		OutputFormat:   req.OutputFormat,
		GeneratedBy:    userID,
	}
	if rp.OutputFormat == "" { rp.OutputFormat = "pdf" }
	if rp.Config == "" { rp.Config = "{}" }
	if rp.Filters == "" { rp.Filters = "{}" }
	if rp.Schedule != "" {
		rp.Status = "scheduled"
	}
	if err := s.repo.CreateReport(ctx, rp); err != nil {
		return nil, err
	}
	return rp, nil
}

func (s *ReportService) ListReports(ctx context.Context, page, pageSize int) ([]model.Report, int, error) {
	return s.repo.ListReports(ctx, page, pageSize)
}

func (s *ReportService) GetReport(ctx context.Context, id int64) (*model.Report, error) {
	return s.repo.GetReport(ctx, id)
}

func (s *ReportService) UpdateReport(ctx context.Context, id int64, req *model.ReportCreate) (*model.Report, error) {
	rp, err := s.repo.GetReport(ctx, id)
	if err != nil { return nil, err }
	rp.Name = req.Name
	rp.Config = req.Config
	rp.Schedule = req.Schedule
	rp.Filters = req.Filters
	if err := s.repo.UpdateReport(ctx, rp); err != nil {
		return nil, err
	}
	return rp, nil
}

func (s *ReportService) DeleteReport(ctx context.Context, id int64) error {
	return s.repo.DeleteReport(ctx, id)
}

func (s *ReportService) Generate(ctx context.Context, id int64) (*model.Report, error) {
	rp, err := s.repo.GetReport(ctx, id)
	if err != nil { return nil, err }
	rp.Status = "generating"
	_ = s.repo.UpdateReport(ctx, rp)

	// Simulate report generation
	now := time.Now()
	fileName := fmt.Sprintf("report_%d_%s.%s", id, now.Format("20060102150405"), rp.OutputFormat)
	rp.FilePath = "/exports/" + fileName
	rp.Status = "completed"
	generatedAt := now
	rp.GeneratedAt = &generatedAt
	_ = s.repo.UpdateReport(ctx, rp)
	return rp, nil
}
