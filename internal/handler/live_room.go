package handler

import (
	"strconv"

	"github.com/CodingFervor/live-commerce-bi/internal/model"
	"github.com/CodingFervor/live-commerce-bi/internal/service"
	"github.com/CodingFervor/live-commerce-bi/pkg/response"

	"github.com/gin-gonic/gin"
)

type LiveRoomHandler struct {
	svc *service.LiveRoomService
}

func NewLiveRoomHandler() *LiveRoomHandler {
	return &LiveRoomHandler{svc: service.NewLiveRoomService()}
}

func (h *LiveRoomHandler) Create(c *gin.Context) {
	var req model.LiveRoomCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	lr, err := h.svc.Create(c.Request.Context(), &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Created(c, lr)
}

func (h *LiveRoomHandler) List(c *gin.Context) {
	page, pageSize := getPagination(c)
	filters := map[string]string{
		"platform":    c.Query("platform"),
		"status":      c.Query("status"),
		"streamer_id": c.Query("streamer_id"),
		"start_date":  c.Query("start_date"),
		"end_date":    c.Query("end_date"),
		"search":      c.Query("search"),
	}
	list, total, err := h.svc.List(c.Request.Context(), page, pageSize, filters)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.PageOK(c, list, int64(total), page, pageSize)
}

func (h *LiveRoomHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	lr, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "live room not found")
		return
	}
	response.OK(c, lr)
}

func (h *LiveRoomHandler) GetMetrics(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	metrics, err := h.svc.GetMetrics(c.Request.Context(), id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, metrics)
}

func (h *LiveRoomHandler) GetProducts(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	products, err := h.svc.GetProducts(c.Request.Context(), id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, products)
}

func (h *LiveRoomHandler) GetFunnel(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	funnel, err := h.svc.GetFunnel(c.Request.Context(), id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, funnel)
}
