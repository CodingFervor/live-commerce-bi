package handler

import (
	"strconv"

	"github.com/CodingFervor/live-commerce-bi/internal/service"
	"github.com/CodingFervor/live-commerce-bi/pkg/response"

	"github.com/gin-gonic/gin"
)

// ═══ Data Screen Handler ═══
// Big-screen display APIs for ops war room

type DataScreenHandler struct {
	svc *service.DataScreenService
}

func NewDataScreenHandler() *DataScreenHandler {
	return &DataScreenHandler{svc: service.NewDataScreenService()}
}

// Overview returns big-screen KPI overview cards
// GET /api/v1/screen/overview
func (h *DataScreenHandler) Overview(c *gin.Context) {
	data, err := h.svc.ScreenOverview(c.Request.Context())
	if err != nil {
		response.InternalError(c, "failed to get screen overview: "+err.Error())
		return
	}
	response.OK(c, data)
}

// Realtime returns live room real-time metrics
// GET /api/v1/screen/realtime
func (h *DataScreenHandler) Realtime(c *gin.Context) {
	data, err := h.svc.ScreenRealtime(c.Request.Context())
	if err != nil {
		response.InternalError(c, "failed to get realtime data: "+err.Error())
		return
	}
	response.OK(c, data)
}

// Rankings returns various rankings for big-screen display
// GET /api/v1/screen/rankings?type=streamer_gmv&period=today&limit=10
func (h *DataScreenHandler) Rankings(c *gin.Context) {
	rankingType := c.DefaultQuery("type", "streamer_gmv")
	period := c.DefaultQuery("period", "today")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	data, err := h.svc.ScreenRankings(c.Request.Context(), rankingType, period, limit)
	if err != nil {
		response.InternalError(c, "failed to get rankings: "+err.Error())
		return
	}
	response.OK(c, data)
}

// Geographic returns geographic distribution
// GET /api/v1/screen/geographic
func (h *DataScreenHandler) Geographic(c *gin.Context) {
	data, err := h.svc.ScreenGeographic(c.Request.Context())
	if err != nil {
		response.InternalError(c, "failed to get geographic data: "+err.Error())
		return
	}
	response.OK(c, data)
}

// ═══ Alert Engine Handler ═══
// Manual trigger and status for intelligent alert engine

type AlertEngineHandler struct {
	engine *service.AlertEngine
}

func NewAlertEngineHandler() *AlertEngineHandler {
	return &AlertEngineHandler{engine: service.NewAlertEngine()}
}

// EvaluateAll manually triggers evaluation of all rules
// POST /api/v1/alerts/evaluate
func (h *AlertEngineHandler) EvaluateAll(c *gin.Context) {
	results, err := h.engine.EvaluateAll(c.Request.Context())
	if err != nil {
		response.InternalError(c, "alert evaluation failed: "+err.Error())
		return
	}
	response.OK(c, results)
}

// EvaluateRule manually evaluates a single rule
// POST /api/v1/alerts/:id/evaluate
func (h *AlertEngineHandler) EvaluateRule(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	result, err := h.engine.EvaluateRule(c.Request.Context(), id)
	if err != nil {
		response.InternalError(c, "alert evaluation failed: "+err.Error())
		return
	}
	response.OK(c, result)
}
