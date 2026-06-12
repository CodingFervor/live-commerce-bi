package response

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// ═══ Enhanced Response Package ═══
// Unified error handling with request tracing

type Response struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
}

type PageData struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, newResponse(c, 0, "success", data))
}

func OKMsg(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, newResponse(c, 0, msg, nil))
}

func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, newResponse(c, 0, "created", data))
}

func PageOK(c *gin.Context, list interface{}, total int64, page, pageSize int) {
	OK(c, PageData{List: list, Total: total, Page: page, PageSize: pageSize})
}

func Error(c *gin.Context, httpCode int, msg string) {
	c.JSON(httpCode, newResponse(c, -1, msg, nil))
}

func BadRequest(c *gin.Context, msg string) {
	Error(c, http.StatusBadRequest, msg)
}

func Unauthorized(c *gin.Context, msg string) {
	Error(c, http.StatusUnauthorized, msg)
}

func Forbidden(c *gin.Context, msg string) {
	Error(c, http.StatusForbidden, msg)
}

func NotFound(c *gin.Context, msg string) {
	Error(c, http.StatusNotFound, msg)
}

func InternalError(c *gin.Context, msg string) {
	Error(c, http.StatusInternalServerError, msg)
}

func TooManyRequests(c *gin.Context, msg string) {
	Error(c, http.StatusTooManyRequests, msg)
}

func ServiceUnavailable(c *gin.Context, msg string) {
	Error(c, http.StatusServiceUnavailable, msg)
}

// newResponse creates a response with request ID
func newResponse(c *gin.Context, code int, msg string, data interface{}) Response {
	r := Response{
		Code:    code,
		Message: msg,
		Data:    data,
	}
	if c != nil {
		r.RequestID = c.GetString("request_id")
	}
	return r
}

// RecoveryHandler returns a gin.HandlerFunc that recovers from panics
// and returns a proper JSON error response with stack trace in dev mode
func RecoveryHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the stack trace
				debug.PrintStack()

				// Return JSON error
				c.AbortWithStatusJSON(http.StatusInternalServerError, Response{
					Code:      -1,
					Message:   "internal server error",
					RequestID: c.GetString("request_id"),
				})
			}
		}()
		c.Next()
	}
}
