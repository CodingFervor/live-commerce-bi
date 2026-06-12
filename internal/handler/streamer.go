package handler

import (
	"strconv"

	"github.com/CodingFervor/live-commerce-bi/internal/model"
	"github.com/CodingFervor/live-commerce-bi/internal/service"
	"github.com/CodingFervor/live-commerce-bi/pkg/response"

	"github.com/gin-gonic/gin"
)

type StreamerHandler struct {
	svc *service.StreamerService
}

func NewStreamerHandler() *StreamerHandler {
	return &StreamerHandler{svc: service.NewStreamerService()}
}

func (h *StreamerHandler) Create(c *gin.Context) {
	var req model.StreamerCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	s, err := h.svc.Create(c.Request.Context(), &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Created(c, s)
}

func (h *StreamerHandler) List(c *gin.Context) {
	page, pageSize := getPagination(c)
	list, total, err := h.svc.List(c.Request.Context(), page, pageSize, c.Query("platform"), c.Query("category"))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.PageOK(c, list, int64(total), page, pageSize)
}

func (h *StreamerHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	s, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "streamer not found")
		return
	}
	response.OK(c, s)
}

func (h *StreamerHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req model.StreamerCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	s, err := h.svc.Update(c.Request.Context(), id, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, s)
}

func (h *StreamerHandler) Rankings(c *gin.Context) {
	metric := c.DefaultQuery("metric", "gmv")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	rankings, err := h.svc.GetRankings(c.Request.Context(), metric, limit, c.Query("platform"))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, rankings)
}

func (h *StreamerHandler) Performance(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	perf, err := h.svc.GetPerformance(c.Request.Context(), id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, perf)
}
