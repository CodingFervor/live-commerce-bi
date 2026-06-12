package handler

import (
	"strconv"

	"github.com/CodingFervor/live-commerce-bi/internal/middleware"
	"github.com/CodingFervor/live-commerce-bi/internal/model"
	"github.com/CodingFervor/live-commerce-bi/internal/service"
	"github.com/CodingFervor/live-commerce-bi/pkg/response"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	svc         *service.DashboardService
	analyticsSvc *service.AnalyticsService
	liveRoomSvc  *service.LiveRoomService
}

func NewDashboardHandler() *DashboardHandler {
	return &DashboardHandler{
		svc:          service.NewDashboardService(),
		analyticsSvc: service.NewAnalyticsService(),
		liveRoomSvc:  service.NewLiveRoomService(),
	}
}

func (h *DashboardHandler) Overview(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	stats, err := h.analyticsSvc.GetOverview(c.Request.Context(), startDate, endDate)
	if err != nil {
		response.InternalError(c, "failed to get overview: "+err.Error())
		return
	}
	response.OK(c, stats)
}

func (h *DashboardHandler) Realtime(c *gin.Context) {
	rooms, err := h.liveRoomSvc.GetActiveRooms(c.Request.Context())
	if err != nil {
		response.InternalError(c, "failed to get realtime data")
		return
	}
	response.OK(c, gin.H{
		"active_rooms": len(rooms),
		"rooms":        rooms,
	})
}

func (h *DashboardHandler) Trend(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	trend, err := h.analyticsSvc.GetGMVTrend(c.Request.Context(), startDate, endDate)
	if err != nil {
		response.InternalError(c, "failed to get trend data")
		return
	}
	response.OK(c, trend)
}

func (h *DashboardHandler) CreateDashboard(c *gin.Context) {
	var req model.DashboardCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	userID := middleware.GetUserID(c)
	d, err := h.svc.Create(c.Request.Context(), &req, userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Created(c, d)
}

func (h *DashboardHandler) ListDashboards(c *gin.Context) {
	userID := middleware.GetUserID(c)
	list, err := h.svc.List(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, list)
}

func (h *DashboardHandler) GetDashboard(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	d, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "dashboard not found")
		return
	}
	response.OK(c, d)
}

func (h *DashboardHandler) UpdateDashboard(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req model.DashboardCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	d, err := h.svc.Update(c.Request.Context(), id, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, d)
}

func (h *DashboardHandler) DeleteDashboard(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OKMsg(c, "deleted")
}

func (h *DashboardHandler) CreateWidget(c *gin.Context) {
	did, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req model.WidgetCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	w, err := h.svc.CreateWidget(c.Request.Context(), did, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Created(c, w)
}

func (h *DashboardHandler) ListWidgets(c *gin.Context) {
	did, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	list, err := h.svc.ListWidgets(c.Request.Context(), did)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, list)
}

func (h *DashboardHandler) UpdateWidget(c *gin.Context) {
	_, _ = strconv.ParseInt(c.Param("id"), 10, 64)
	wid, _ := strconv.ParseInt(c.Param("wid"), 10, 64)
	var req model.WidgetCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	w, err := h.svc.UpdateWidget(c.Request.Context(), wid, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, w)
}

func (h *DashboardHandler) DeleteWidget(c *gin.Context) {
	wid, _ := strconv.ParseInt(c.Param("wid"), 10, 64)
	if err := h.svc.DeleteWidget(c.Request.Context(), wid); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OKMsg(c, "deleted")
}
