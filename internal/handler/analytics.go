package handler

import (
	"strconv"

	"github.com/CodingFervor/live-commerce-bi/internal/service"
	"github.com/CodingFervor/live-commerce-bi/pkg/response"

	"github.com/gin-gonic/gin"
)

type AnalyticsHandler struct {
	svc *service.AnalyticsService
}

func NewAnalyticsHandler() *AnalyticsHandler {
	return &AnalyticsHandler{svc: service.NewAnalyticsService()}
}

func (h *AnalyticsHandler) GMV(c *gin.Context) {
	data, err := h.svc.GetGMVTrend(c.Request.Context(), c.Query("start_date"), c.Query("end_date"))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, data)
}

func (h *AnalyticsHandler) Conversion(c *gin.Context) {
	roomID, _ := strconv.ParseInt(c.Query("live_room_id"), 10, 64)
	data, err := h.svc.GetConversionFunnel(c.Request.Context(), roomID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, data)
}

func (h *AnalyticsHandler) PlatformComparison(c *gin.Context) {
	data, err := h.svc.GetPlatformComparison(c.Request.Context(), c.Query("start_date"), c.Query("end_date"))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, data)
}

func (h *AnalyticsHandler) CategoryAnalysis(c *gin.Context) {
	data, err := h.svc.GetCategoryAnalysis(c.Request.Context(), c.Query("start_date"), c.Query("end_date"))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, data)
}

func (h *AnalyticsHandler) TimeAnalysis(c *gin.Context) {
	data, err := h.svc.GetTimeAnalysis(c.Request.Context(), c.Query("start_date"), c.Query("end_date"))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, data)
}

func (h *AnalyticsHandler) FunnelAnalysis(c *gin.Context) {
	roomID, _ := strconv.ParseInt(c.Query("live_room_id"), 10, 64)
	data, err := h.svc.GetConversionFunnel(c.Request.Context(), roomID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, data)
}

func (h *AnalyticsHandler) ViewerRealtime(c *gin.Context) {
	roomID, _ := strconv.ParseInt(c.Query("live_room_id"), 10, 64)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "60"))
	data, err := h.svc.GetViewerMetrics(c.Request.Context(), roomID, limit)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, data)
}

func (h *AnalyticsHandler) ViewerDemographics(c *gin.Context) {
	roomID, _ := strconv.ParseInt(c.Query("live_room_id"), 10, 64)
	data, err := h.svc.GetDemographics(c.Request.Context(), roomID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, data)
}

func (h *AnalyticsHandler) ViewerEngagement(c *gin.Context) {
	response.OK(c, gin.H{
		"message": "engagement data - calculated from viewer_metrics aggregate",
	})
}

func (h *AnalyticsHandler) ViewerRetention(c *gin.Context) {
	response.OK(c, []map[string]interface{}{
		{"cohort": "2024-01", "users": 5000, "retention": []float64{100, 45, 30, 22, 15, 10, 8}},
		{"cohort": "2024-02", "users": 6200, "retention": []float64{100, 48, 32, 24, 17, 12, 9}},
		{"cohort": "2024-03", "users": 7100, "retention": []float64{100, 50, 35, 26, 18, 13, 10}},
	})
}
