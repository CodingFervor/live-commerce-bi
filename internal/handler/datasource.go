package handler

import (
	"net/http"
	"strconv"

	"github.com/CodingFervor/live-commerce-bi/internal/middleware"
	"github.com/CodingFervor/live-commerce-bi/internal/model"
	"github.com/CodingFervor/live-commerce-bi/internal/service"
	"github.com/CodingFervor/live-commerce-bi/pkg/response"

	"github.com/gin-gonic/gin"
)

type DataSourceHandler struct {
	svc *service.DataSourceService
}

func NewDataSourceHandler() *DataSourceHandler {
	return &DataSourceHandler{svc: service.NewDataSourceService()}
}

func (h *DataSourceHandler) Create(c *gin.Context) {
	var req model.DataSourceCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	ds, err := h.svc.Create(c.Request.Context(), &req, middleware.GetUserID(c))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Created(c, ds)
}

func (h *DataSourceHandler) List(c *gin.Context) {
	page, pageSize := getPagination(c)
	list, total, err := h.svc.List(c.Request.Context(), page, pageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.PageOK(c, list, int64(total), page, pageSize)
}

func (h *DataSourceHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	ds, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "data source not found")
		return
	}
	response.OK(c, ds)
}

func (h *DataSourceHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req model.DataSourceUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	ds, err := h.svc.Update(c.Request.Context(), id, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, ds)
}

func (h *DataSourceHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OKMsg(c, "deleted")
}

func (h *DataSourceHandler) Sync(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.svc.TriggerSync(c.Request.Context(), id); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OKMsg(c, "sync triggered")
}

func (h *DataSourceHandler) Status(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	status, err := h.svc.GetStatus(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "data source not found")
		return
	}
	response.OK(c, status)
}

func getPagination(c *gin.Context) (int, int) {
	page := 1
	pageSize := 20
	if p := c.Query("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 && v <= 100 {
			pageSize = v
		}
	}
	return page, pageSize
}

// suppress unused import
var _ = http.StatusOK
