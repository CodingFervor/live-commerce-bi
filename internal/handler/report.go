package handler

import (
	"path/filepath"
	"strconv"
	"strings"

	"github.com/CodingFervor/live-commerce-bi/internal/middleware"
	"github.com/CodingFervor/live-commerce-bi/internal/model"
	"github.com/CodingFervor/live-commerce-bi/internal/service"
	"github.com/CodingFervor/live-commerce-bi/pkg/response"

	"github.com/gin-gonic/gin"
)

type ReportHandler struct {
	svc *service.ReportService
}

func NewReportHandler() *ReportHandler {
	return &ReportHandler{svc: service.NewReportService()}
}

func (h *ReportHandler) Create(c *gin.Context) {
	var req model.ReportCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	rp, err := h.svc.CreateReport(c.Request.Context(), &req, middleware.GetUserID(c))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Created(c, rp)
}

func (h *ReportHandler) List(c *gin.Context) {
	page, pageSize := getPagination(c)
	list, total, err := h.svc.ListReports(c.Request.Context(), page, pageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.PageOK(c, list, int64(total), page, pageSize)
}

func (h *ReportHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	rp, err := h.svc.GetReport(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "report not found")
		return
	}
	// Ownership check
	userID := middleware.GetUserID(c)
	role, _ := c.Get("role")
	if role != "admin" && rp.GeneratedBy != userID {
		response.Forbidden(c, "access denied")
		return
	}
	response.OK(c, rp)
}

func (h *ReportHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	// Ownership check
	existing, err := h.svc.GetReport(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "report not found")
		return
	}
	userID := middleware.GetUserID(c)
	role, _ := c.Get("role")
	if role != "admin" && existing.GeneratedBy != userID {
		response.Forbidden(c, "access denied")
		return
	}
	var req model.ReportCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	rp, err := h.svc.UpdateReport(c.Request.Context(), id, &req)
	if err != nil {
		response.InternalError(c, "failed to update report")
		return
	}
	response.OK(c, rp)
}

func (h *ReportHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	// Ownership check
	existing, err := h.svc.GetReport(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "report not found")
		return
	}
	userID := middleware.GetUserID(c)
	role, _ := c.Get("role")
	if role != "admin" && existing.GeneratedBy != userID {
		response.Forbidden(c, "access denied")
		return
	}
	if err := h.svc.DeleteReport(c.Request.Context(), id); err != nil {
		response.InternalError(c, "failed to delete report")
		return
	}
	response.OKMsg(c, "deleted")
}

func (h *ReportHandler) Generate(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	rp, err := h.svc.Generate(c.Request.Context(), id)
	if err != nil {
		response.InternalError(c, "failed to generate report")
		return
	}
	response.OK(c, rp)
}

func (h *ReportHandler) Download(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	rp, err := h.svc.GetReport(c.Request.Context(), id)
	if err != nil || rp.FilePath == "" {
		response.NotFound(c, "report file not found")
		return
	}
	// Ownership check
	userID := middleware.GetUserID(c)
	role, _ := c.Get("role")
	if role != "admin" && rp.GeneratedBy != userID {
		response.Forbidden(c, "access denied")
		return
	}
	// Prevent path traversal: ensure file path is within allowed directory
	absPath, err := filepath.Abs(rp.FilePath)
	if err != nil || !strings.HasPrefix(absPath, "reports") && !strings.HasPrefix(absPath, "exports") {
		response.InternalError(c, "invalid file path")
		return
	}
	c.File(absPath)
}

func (h *ReportHandler) ListTemplates(c *gin.Context) {
	list, err := h.svc.ListTemplates(c.Request.Context())
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, list)
}

func (h *ReportHandler) CreateTemplate(c *gin.Context) {
	var t model.ReportTemplate
	if err := c.ShouldBindJSON(&t); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	tmpl, err := h.svc.CreateTemplate(c.Request.Context(), &t, middleware.GetUserID(c))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Created(c, tmpl)
}
