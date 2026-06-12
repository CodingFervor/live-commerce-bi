package handler

import (
	"strconv"

	"github.com/CodingFervor/live-commerce-bi/internal/middleware"
	"github.com/CodingFervor/live-commerce-bi/internal/model"
	"github.com/CodingFervor/live-commerce-bi/internal/service"
	"github.com/CodingFervor/live-commerce-bi/pkg/response"

	"github.com/gin-gonic/gin"
)

type AlertHandler struct {
	svc *service.AlertService
}

func NewAlertHandler() *AlertHandler {
	return &AlertHandler{svc: service.NewAlertService()}
}

func (h *AlertHandler) CreateRule(c *gin.Context) {
	var req model.AlertRuleCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	rule, err := h.svc.CreateRule(c.Request.Context(), &req, middleware.GetUserID(c))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Created(c, rule)
}

func (h *AlertHandler) ListRules(c *gin.Context) {
	list, err := h.svc.ListRules(c.Request.Context())
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, list)
}

func (h *AlertHandler) GetRule(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	rule, err := h.svc.GetRule(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "rule not found")
		return
	}
	response.OK(c, rule)
}

func (h *AlertHandler) UpdateRule(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req model.AlertRuleCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	rule, err := h.svc.UpdateRule(c.Request.Context(), id, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, rule)
}

func (h *AlertHandler) DeleteRule(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.svc.DeleteRule(c.Request.Context(), id); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OKMsg(c, "deleted")
}

func (h *AlertHandler) GetHistory(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	page, pageSize := getPagination(c)
	list, total, err := h.svc.GetHistory(c.Request.Context(), id, page, pageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.PageOK(c, list, int64(total), page, pageSize)
}

func (h *AlertHandler) TestAlert(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.svc.TestAlert(c.Request.Context(), id); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OKMsg(c, "test alert sent")
}
