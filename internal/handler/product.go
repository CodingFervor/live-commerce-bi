package handler

import (
	"strconv"

	"github.com/CodingFervor/live-commerce-bi/internal/model"
	"github.com/CodingFervor/live-commerce-bi/internal/service"
	"github.com/CodingFervor/live-commerce-bi/pkg/response"

	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	svc *service.ProductService
}

func NewProductHandler() *ProductHandler {
	return &ProductHandler{svc: service.NewProductService()}
}

func (h *ProductHandler) Create(c *gin.Context) {
	var req model.ProductCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	p, err := h.svc.Create(c.Request.Context(), &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Created(c, p)
}

func (h *ProductHandler) List(c *gin.Context) {
	page, pageSize := getPagination(c)
	list, total, err := h.svc.List(c.Request.Context(), page, pageSize, c.Query("platform"), c.Query("category"))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.PageOK(c, list, int64(total), page, pageSize)
}

func (h *ProductHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	p, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "product not found")
		return
	}
	response.OK(c, p)
}

func (h *ProductHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req model.ProductCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	p, err := h.svc.Update(c.Request.Context(), id, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, p)
}

func (h *ProductHandler) Rankings(c *gin.Context) {
	metric := c.DefaultQuery("metric", "revenue")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	rankings, err := h.svc.GetRankings(c.Request.Context(), metric, limit, c.Query("category"))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, rankings)
}
