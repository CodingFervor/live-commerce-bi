package handler

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/CodingFervor/live-commerce-bi/internal/middleware"
	"github.com/CodingFervor/live-commerce-bi/internal/model"
	"github.com/CodingFervor/live-commerce-bi/internal/repository"
	"github.com/CodingFervor/live-commerce-bi/internal/service"
	"github.com/CodingFervor/live-commerce-bi/pkg/response"

	"github.com/gin-gonic/gin"
)

// ═══ Douyin Compass Handler ═══

type CompassHandler struct {
	repo   *repository.CompassRepo
	engine *service.CompassEngine
}

func NewCompassHandler() *CompassHandler {
	return &CompassHandler{
		repo:   repository.NewCompassRepo(),
		engine: service.NewCompassEngine(),
	}
}

// ─── Session Management ───

func (h *CompassHandler) CreateSession(c *gin.Context) {
	var req model.CompassSessionCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	sess := &model.CompassSession{
		Name:         req.Name,
		Cookie:       req.Cookie,
		ShopID:       req.ShopID,
		ShopName:     req.ShopName,
		UserAgent:    req.UserAgent,
		ProxyURL:     req.ProxyURL,
		MaxDailyReqs: req.MaxDailyReqs,
		CreatedBy:    middleware.GetUserID(c),
	}
	if sess.MaxDailyReqs == 0 {
		sess.MaxDailyReqs = 300 // safe default: 300 requests/day
	}

	if err := h.repo.CreateSession(c.Request.Context(), sess); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	// Mask cookie in response
	sess.Cookie = "***"
	response.Created(c, sess)
}

func (h *CompassHandler) ListSessions(c *gin.Context) {
	page, pageSize := getPagination(c)
	list, total, err := h.repo.ListSessions(c.Request.Context(), page, pageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	// Mask cookies in response
	for i := range list {
		list[i].Cookie = "***"
	}
	response.PageOK(c, list, int64(total), page, pageSize)
}

func (h *CompassHandler) GetSession(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	sess, err := h.repo.GetSession(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "session not found")
		return
	}
	sess.Cookie = "***"
	response.OK(c, sess)
}

func (h *CompassHandler) UpdateSession(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req model.CompassSessionUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	sess, err := h.repo.GetSession(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "session not found")
		return
	}

	if req.Name != nil {
		sess.Name = *req.Name
	}
	if req.Cookie != nil {
		sess.Cookie = *req.Cookie
		h.engine.InvalidateSession(id) // force re-create HTTP client
	}
	if req.UserAgent != nil {
		sess.UserAgent = *req.UserAgent
	}
	if req.ProxyURL != nil {
		sess.ProxyURL = *req.ProxyURL
	}
	if req.Status != nil {
		sess.Status = *req.Status
	}
	if req.MaxDailyReqs != nil {
		sess.MaxDailyReqs = *req.MaxDailyReqs
	}

	if err := h.repo.UpdateSession(c.Request.Context(), sess); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	sess.Cookie = "***"
	response.OK(c, sess)
}

func (h *CompassHandler) DeleteSession(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.repo.DeleteSession(c.Request.Context(), id); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	h.engine.InvalidateSession(id)
	response.OKMsg(c, "session deleted")
}

func (h *CompassHandler) CheckSessionHealth(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	status, err := h.engine.CheckSessionHealth(c.Request.Context(), id)
	if err != nil {
		response.OK(c, gin.H{"status": status, "error": err.Error()})
		return
	}
	response.OK(c, gin.H{"status": status})
}

// ─── Task Management ───

func (h *CompassHandler) CreateTask(c *gin.Context) {
	var req model.CompassTaskCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	task := &model.CompassTask{
		SessionID: req.SessionID,
		TaskType:  req.TaskType,
		Params:    req.Params,
	}

	if err := h.repo.CreateTask(c.Request.Context(), task); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Created(c, task)
}

func (h *CompassHandler) ListTasks(c *gin.Context) {
	sessionID, _ := strconv.ParseInt(c.Query("session_id"), 10, 64)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	list, err := h.repo.ListTasks(c.Request.Context(), sessionID, limit)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, list)
}

func (h *CompassHandler) GetTask(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	task, err := h.repo.GetTask(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "task not found")
		return
	}
	response.OK(c, task)
}

// ─── Data Fetching Endpoints ───

func (h *CompassHandler) FetchLiveOverview(c *gin.Context) {
	sessionID, _ := strconv.ParseInt(c.Query("session_id"), 10, 64)
	startDate := c.DefaultQuery("start_date", time.Now().AddDate(0, 0, -7).Format("2006-01-02"))
	endDate := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

	data, err := h.engine.FetchLiveOverview(c.Request.Context(), sessionID, startDate, endDate)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, data)
}

func (h *CompassHandler) FetchLiveDetail(c *gin.Context) {
	sessionID, _ := strconv.ParseInt(c.Query("session_id"), 10, 64)
	roomID := c.Param("room_id")

	data, err := h.engine.FetchLiveDetail(c.Request.Context(), sessionID, roomID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, data)
}

func (h *CompassHandler) FetchProducts(c *gin.Context) {
	sessionID, _ := strconv.ParseInt(c.Query("session_id"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	category := c.Query("category")

	data, err := h.engine.FetchProductList(c.Request.Context(), sessionID, page, 20, category)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, data)
}

func (h *CompassHandler) FetchProductDetail(c *gin.Context) {
	sessionID, _ := strconv.ParseInt(c.Query("session_id"), 10, 64)
	productID := c.Param("product_id")

	data, err := h.engine.FetchProductDetail(c.Request.Context(), sessionID, productID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, data)
}

func (h *CompassHandler) FetchOrders(c *gin.Context) {
	sessionID, _ := strconv.ParseInt(c.Query("session_id"), 10, 64)
	startDate := c.DefaultQuery("start_date", time.Now().AddDate(0, 0, -7).Format("2006-01-02"))
	endDate := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	data, err := h.engine.FetchOrderList(c.Request.Context(), sessionID, startDate, endDate, page, 50)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, data)
}

func (h *CompassHandler) FetchStreamerRank(c *gin.Context) {
	sessionID, _ := strconv.ParseInt(c.Query("session_id"), 10, 64)
	startDate := c.DefaultQuery("start_date", time.Now().AddDate(0, 0, -7).Format("2006-01-02"))
	endDate := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

	data, err := h.engine.FetchStreamerRank(c.Request.Context(), sessionID, startDate, endDate)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, data)
}

func (h *CompassHandler) FetchFunnel(c *gin.Context) {
	sessionID, _ := strconv.ParseInt(c.Query("session_id"), 10, 64)
	startDate := c.DefaultQuery("start_date", time.Now().AddDate(0, 0, -7).Format("2006-01-02"))
	endDate := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

	data, err := h.engine.FetchFunnelAnalysis(c.Request.Context(), sessionID, startDate, endDate)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, data)
}

// RunSync executes a full sync cycle (runs as background task)
func (h *CompassHandler) RunSync(c *gin.Context) {
	var req struct {
		SessionID int64  `json:"session_id" binding:"required"`
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if req.StartDate == "" {
		req.StartDate = time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	}
	if req.EndDate == "" {
		req.EndDate = time.Now().Format("2006-01-02")
	}

	// Create a task record
	params, _ := json.Marshal(map[string]string{
		"start_date": req.StartDate,
		"end_date":   req.EndDate,
	})
	task := &model.CompassTask{
		SessionID: req.SessionID,
		TaskType:  "live_overview", // full sync covers all types
		Params:    string(params),
	}
	if err := h.repo.CreateTask(c.Request.Context(), task); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	// Run sync in background goroutine
	go func() {
		bgCtx := c.Request.Context()
		// Update task status to running
		h.repo.UpdateTaskStatus(bgCtx, task.ID, "running", "", "", 0)

		result, err := h.engine.RunFullSync(bgCtx, req.SessionID, req.StartDate, req.EndDate)
		if err != nil {
			h.repo.UpdateTaskStatus(bgCtx, task.ID, "failed", "", err.Error(), 0)
			return
		}

		resultJSON, _ := json.Marshal(map[string]interface{}{
			"total_requests": result.TotalRequests,
			"duration":       result.Duration,
			"live_count":     len(result.LiveOverview),
			"product_count":  len(result.Products),
			"order_count":    len(result.Orders),
			"streamer_count": len(result.Streamers),
			"funnel_count":   len(result.Funnel),
			"errors":         result.Errors,
		})

		totalRecords := len(result.LiveOverview) + len(result.Products) +
			len(result.Orders) + len(result.Streamers) + len(result.Funnel)

		status := "completed"
		errMsg := ""
		if len(result.Errors) > 0 {
			errMsg = result.Errors[0]
		}
		h.repo.UpdateTaskStatus(bgCtx, task.ID, status, string(resultJSON), errMsg, totalRecords)
	}()

	response.OK(c, gin.H{
		"task_id":    task.ID,
		"session_id": req.SessionID,
		"status":     "running",
		"message":    "sync started in background",
	})
}

// suppress unused import
var _ = strconv.Itoa(0)
