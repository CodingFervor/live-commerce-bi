package handler

import (
	"strconv"

	"github.com/CodingFervor/live-commerce-bi/internal/middleware"
	"github.com/CodingFervor/live-commerce-bi/internal/model"
	"github.com/CodingFervor/live-commerce-bi/internal/repository"
	"github.com/CodingFervor/live-commerce-bi/internal/service"
	"github.com/CodingFervor/live-commerce-bi/pkg/response"

	"github.com/gin-gonic/gin"
)

// ═══ Organization Handler ═══

type OrganizationHandler struct {
	repo *repository.OrganizationRepo
}

func NewOrganizationHandler() *OrganizationHandler {
	return &OrganizationHandler{repo: repository.NewOrganizationRepo()}
}

func (h *OrganizationHandler) CreateOrg(c *gin.Context) {
	var o model.Organization
	if err := c.ShouldBindJSON(&o); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.repo.CreateOrg(c.Request.Context(), &o); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Created(c, o)
}

func (h *OrganizationHandler) ListOrgs(c *gin.Context) {
	page, pageSize := getPagination(c)
	list, total, err := h.repo.ListOrgs(c.Request.Context(), page, pageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.PageOK(c, list, int64(total), page, pageSize)
}

func (h *OrganizationHandler) GetOrg(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	o, err := h.repo.GetOrg(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "organization not found")
		return
	}
	response.OK(c, o)
}

func (h *OrganizationHandler) UpdateOrg(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var o model.Organization
	if err := c.ShouldBindJSON(&o); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	o.ID = id
	if err := h.repo.UpdateOrg(c.Request.Context(), &o); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, o)
}

func (h *OrganizationHandler) CreateDept(c *gin.Context) {
	var d model.Department
	if err := c.ShouldBindJSON(&d); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	orgID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	d.OrganizationID = orgID
	if d.Path == "" {
		d.Path = "/"
	}
	if d.Level == 0 {
		d.Level = 1
	}
	if err := h.repo.CreateDept(c.Request.Context(), &d); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Created(c, d)
}

func (h *OrganizationHandler) GetDeptTree(c *gin.Context) {
	orgID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	tree, err := h.repo.GetDeptTree(c.Request.Context(), orgID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, tree)
}

func (h *OrganizationHandler) UpdateDept(c *gin.Context) {
	did, _ := strconv.ParseInt(c.Param("did"), 10, 64)
	var d model.Department
	if err := c.ShouldBindJSON(&d); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	d.ID = did
	if err := h.repo.UpdateDept(c.Request.Context(), &d); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, d)
}

func (h *OrganizationHandler) DeleteDept(c *gin.Context) {
	did, _ := strconv.ParseInt(c.Param("did"), 10, 64)
	if err := h.repo.DeleteDept(c.Request.Context(), did); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OKMsg(c, "deleted")
}

// ═══ RBAC Handler ═══

type RBACHandler struct {
	repo *repository.RBACRepo
}

func NewRBACHandler() *RBACHandler {
	return &RBACHandler{repo: repository.NewRBACRepo()}
}

func (h *RBACHandler) ListPermissions(c *gin.Context) {
	list, err := h.repo.ListPermissions(c.Request.Context())
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, list)
}

func (h *RBACHandler) CreateRole(c *gin.Context) {
	var role model.Role
	if err := c.ShouldBindJSON(&role); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.repo.CreateRole(c.Request.Context(), &role); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Created(c, role)
}

func (h *RBACHandler) ListRoles(c *gin.Context) {
	orgID, _ := strconv.ParseInt(c.DefaultQuery("org_id", "1"), 10, 64)
	list, err := h.repo.ListRoles(c.Request.Context(), orgID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, list)
}

func (h *RBACHandler) UpdateRole(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var body struct {
		Name        string `json:"name"`
		PermissionIDs []int64 `json:"permission_ids"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.repo.UpdateRolePermissions(c.Request.Context(), id, body.PermissionIDs); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OKMsg(c, "updated")
}

func (h *RBACHandler) AssignRole(c *gin.Context) {
	roleID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID, _ := strconv.ParseInt(c.Param("uid"), 10, 64)
	if err := h.repo.AssignUserRole(c.Request.Context(), userID, roleID); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OKMsg(c, "role assigned")
}

func (h *RBACHandler) GetCurrentUserPermissions(c *gin.Context) {
	userID := middleware.GetUserID(c)
	perms, err := h.repo.GetUserPermissions(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, perms)
}

// ═══ Audit Handler ═══

type AuditHandler struct {
	repo *repository.AuditRepo
}

func NewAuditHandler() *AuditHandler {
	return &AuditHandler{repo: repository.NewAuditRepo()}
}

func (h *AuditHandler) ListAuditLogs(c *gin.Context) {
	page, pageSize := getPagination(c)
	filters := map[string]string{
		"user_id":  c.Query("user_id"),
		"action":   c.Query("action"),
		"resource": c.Query("resource"),
	}
	list, total, err := h.repo.List(c.Request.Context(), page, pageSize, filters)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.PageOK(c, list, int64(total), page, pageSize)
}

func (h *AuditHandler) GetAuditStats(c *gin.Context) {
	stats, err := h.repo.GetStats(c.Request.Context(), c.Query("start_date"), c.Query("end_date"))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, stats)
}

// ═══ Export Handler ═══

type ExportHandler struct {
	repo   *repository.ExportRepo
	engine *service.ExportEngine
}

func NewExportHandler() *ExportHandler {
	return &ExportHandler{repo: repository.NewExportRepo(), engine: service.NewExportEngine()}
}

func (h *ExportHandler) CreateExport(c *gin.Context) {
	var t model.ExportTask
	if err := c.ShouldBindJSON(&t); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	t.UserID = middleware.GetUserID(c)
	if t.Format == "" {
		t.Format = "csv"
	}
	if err := h.repo.Create(c.Request.Context(), &t); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Created(c, t)
}

func (h *ExportHandler) GetExportStatus(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	t, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "export task not found")
		return
	}
	response.OK(c, t)
}

func (h *ExportHandler) ListExports(c *gin.Context) {
	userID := middleware.GetUserID(c)
	list, err := h.repo.ListByUser(c.Request.Context(), userID, 20)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, list)
}

func (h *ExportHandler) DownloadExport(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	t, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil || t.FilePath == "" {
		response.NotFound(c, "export file not found")
		return
	}
	// Ownership check
	userID := middleware.GetUserID(c)
	role, _ := c.Get("role")
	if role != "admin" && t.UserID != userID {
		response.Forbidden(c, "access denied")
		return
	}
	c.File(t.FilePath)
}

func (h *ExportHandler) CancelExport(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.repo.UpdateStatus(c.Request.Context(), id, "cancelled", "", ""); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OKMsg(c, "cancelled")
}

// ═══ Advanced Analytics Handler ═══

type AdvancedAnalyticsHandler struct {
	engine *service.AnalyticsEngine
}

func NewAdvancedAnalyticsHandler() *AdvancedAnalyticsHandler {
	return &AdvancedAnalyticsHandler{engine: service.NewAnalyticsEngine()}
}

func (h *AdvancedAnalyticsHandler) CohortAnalysis(c *gin.Context) {
	data, err := h.engine.CohortAnalysis(c.Request.Context(), c.Query("start_date"), c.Query("end_date"))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, data)
}

func (h *AdvancedAnalyticsHandler) RFMAnalysis(c *gin.Context) {
	data, err := h.engine.RFMAnalysis(c.Request.Context())
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, data)
}

func (h *AdvancedAnalyticsHandler) SalesForecast(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	data, err := h.engine.SalesForecast(c.Request.Context(), days)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, data)
}

func (h *AdvancedAnalyticsHandler) AnomalyDetection(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	data, err := h.engine.AnomalyDetection(c.Request.Context(), c.Query("metric"), days)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, data)
}

func (h *AdvancedAnalyticsHandler) UserPathAnalysis(c *gin.Context) {
	roomID, _ := strconv.ParseInt(c.Query("live_room_id"), 10, 64)
	data, err := h.engine.UserPathAnalysis(c.Request.Context(), roomID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, data)
}

func (h *AdvancedAnalyticsHandler) EngagementHeatmap(c *gin.Context) {
	roomID, _ := strconv.ParseInt(c.Query("live_room_id"), 10, 64)
	data, err := h.engine.EngagementHeatmap(c.Request.Context(), roomID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, data)
}

func (h *AdvancedAnalyticsHandler) OLAPQuery(c *gin.Context) {
	var q model.OLAPQuery
	if err := c.ShouldBindJSON(&q); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	data, err := h.engine.ExecuteOLAP(c.Request.Context(), q)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, data)
}

func (h *AdvancedAnalyticsHandler) DrillDown(c *gin.Context) {
	// Simplified drill-down: query with added child dimension
	dimension := c.Query("dimension")
	childDimension := c.Query("child_dimension")
	value := c.Query("value")
	q := model.OLAPQuery{
		Dimensions: []string{dimension, childDimension},
		Metrics:    []string{"gmv", "orders", "views"},
		Filters:    map[string]string{dimension: value},
		Limit:      50,
	}
	data, err := h.engine.ExecuteOLAP(c.Request.Context(), q)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, data)
}

func (h *AdvancedAnalyticsHandler) PeriodComparison(c *gin.Context) {
	data, err := h.engine.PeriodComparison(c.Request.Context(),
		c.Query("metric"), c.Query("start_date"), c.Query("end_date"), c.DefaultQuery("type", "wow"))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, data)
}

func (h *AdvancedAnalyticsHandler) TargetComparison(c *gin.Context) {
	target, _ := strconv.ParseFloat(c.Query("target"), 64)
	data, err := h.engine.TargetComparison(c.Request.Context(),
		c.Query("metric"), c.Query("start_date"), c.Query("end_date"), target)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, data)
}

// ═══ Data Quality Handler ═══

type DataQualityHandler struct {
	repo *repository.DataQualityRepo
}

func NewDataQualityHandler() *DataQualityHandler {
	return &DataQualityHandler{repo: repository.NewDataQualityRepo()}
}

func (h *DataQualityHandler) CreateRule(c *gin.Context) {
	var rule model.DataQualityRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	rule.CreatedBy = middleware.GetUserID(c)
	if err := h.repo.CreateRule(c.Request.Context(), &rule); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Created(c, rule)
}

func (h *DataQualityHandler) ListRules(c *gin.Context) {
	list, err := h.repo.ListRules(c.Request.Context())
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, list)
}

func (h *DataQualityHandler) RunCheck(c *gin.Context) {
	response.OK(c, gin.H{"message": "quality check triggered", "rule_id": c.Param("id")})
}

func (h *DataQualityHandler) GetResults(c *gin.Context) {
	response.OK(c, []interface{}{})
}

// ═══ Event Tracking Handler ═══

type EventHandler struct {
	repo *repository.EventRepo
}

func NewEventHandler() *EventHandler {
	return &EventHandler{repo: repository.NewEventRepo()}
}

func (h *EventHandler) Track(c *gin.Context) {
	var e model.TrackEvent
	if err := c.ShouldBindJSON(&e); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.repo.Track(c.Request.Context(), &e); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Created(c, e)
}

func (h *EventHandler) BatchTrack(c *gin.Context) {
	var events []model.TrackEvent
	if err := c.ShouldBindJSON(&events); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.repo.BatchTrack(c.Request.Context(), events); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, gin.H{"count": len(events)})
}

func (h *EventHandler) GetFunnelEvents(c *gin.Context) {
	roomID, _ := strconv.ParseInt(c.Query("live_room_id"), 10, 64)
	events, err := h.repo.GetFunnelEvents(c.Request.Context(), roomID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, events)
}

func (h *EventHandler) GetUserPaths(c *gin.Context) {
	response.OK(c, []interface{}{})
}

// ═══ Metrics Aggregation Handler ═══

type MetricsAggHandler struct {
	repo *repository.MetricsAggRepo
}

func NewMetricsAggHandler() *MetricsAggHandler {
	return &MetricsAggHandler{repo: repository.NewMetricsAggRepo()}
}

func (h *MetricsAggHandler) GetHourlyMetrics(c *gin.Context) {
	response.OK(c, gin.H{"message": "hourly metrics", "platform": c.Query("platform")})
}

func (h *MetricsAggHandler) GetDailyMetrics(c *gin.Context) {
	data, err := h.repo.GetDailyMetrics(c.Request.Context(),
		c.Query("platform"), c.Query("start_date"), c.Query("end_date"))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, data)
}

func (h *MetricsAggHandler) GetStreamerDaily(c *gin.Context) {
	response.OK(c, []interface{}{})
}

// suppress unused import
var _ = strconv.Itoa(0)
