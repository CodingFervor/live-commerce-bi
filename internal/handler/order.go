package handler

import (
	"strconv"

	"github.com/CodingFervor/live-commerce-bi/internal/service"
	"github.com/CodingFervor/live-commerce-bi/pkg/response"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	svc *service.OrderService
}

func NewOrderHandler() *OrderHandler {
	return &OrderHandler{svc: service.NewOrderService()}
}

func (h *OrderHandler) List(c *gin.Context) {
	page, pageSize := getPagination(c)
	filters := map[string]string{
		"platform":     c.Query("platform"),
		"status":       c.Query("status"),
		"live_room_id": c.Query("live_room_id"),
		"streamer_id":  c.Query("streamer_id"),
		"start_date":   c.Query("start_date"),
		"end_date":     c.Query("end_date"),
	}
	list, total, err := h.svc.List(c.Request.Context(), page, pageSize, filters)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.PageOK(c, list, int64(total), page, pageSize)
}

func (h *OrderHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	o, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "order not found")
		return
	}
	response.OK(c, o)
}

func (h *OrderHandler) Stats(c *gin.Context) {
	stats, err := h.svc.GetStats(c.Request.Context(), c.Query("start_date"), c.Query("end_date"))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, stats)
}

func (h *OrderHandler) Revenue(c *gin.Context) {
	revenue, err := h.svc.GetRevenue(c.Request.Context(), c.Query("start_date"), c.Query("end_date"), c.Query("group_by"))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, revenue)
}

// suppress unused
var _ = strconv.Itoa(0)
