package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// ═══ RequestID Middleware Tests ═══

func TestRequestIDGenerates(t *testing.T) {
	r := gin.New()
	r.Use(RequestID())
	r.GET("/test", func(c *gin.Context) {
		rid := c.GetString("request_id")
		if rid == "" {
			t.Error("request_id should not be empty")
		}
		c.String(200, rid)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Status should be 200, got %d", w.Code)
	}

	// Check X-Request-ID header is set
	rid := w.Header().Get("X-Request-ID")
	if rid == "" {
		t.Error("X-Request-ID header should be set")
	}
}

func TestRequestIDPreservesExisting(t *testing.T) {
	r := gin.New()
	r.Use(RequestID())
	r.GET("/test", func(c *gin.Context) {
		rid := c.GetString("request_id")
		c.String(200, rid)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", "custom-request-id-123")
	r.ServeHTTP(w, req)

	body := w.Body.String()
	if body != "custom-request-id-123" {
		t.Errorf("Should preserve existing request ID, got %s", body)
	}

	rid := w.Header().Get("X-Request-ID")
	if rid != "custom-request-id-123" {
		t.Errorf("Header should match, got %s", rid)
	}
}

func TestRequestIDUnique(t *testing.T) {
	r := gin.New()
	r.Use(RequestID())
	r.GET("/test", func(c *gin.Context) {
		c.String(200, c.GetString("request_id"))
	})

	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		r.ServeHTTP(w, req)
		id := w.Body.String()
		if ids[id] {
			t.Errorf("Duplicate request ID: %s", id)
		}
		ids[id] = true
	}
}

// ═══ CORS Middleware Tests ═══

func TestCORSMiddleware(t *testing.T) {
	r := gin.New()
	r.Use(CORS())
	r.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	r.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") == "" {
		t.Error("CORS origin header should be set")
	}
}

func TestCORSPreflight(t *testing.T) {
	r := gin.New()
	r.Use(CORS())
	r.OPTIONS("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "POST")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent && w.Code != http.StatusOK {
		t.Errorf("Preflight should return 204 or 200, got %d", w.Code)
	}
}

// ═══ Auth Middleware Tests ═══

func TestAuthRequiredNoToken(t *testing.T) {
	r := gin.New()
	r.Use(AuthRequired())
	r.GET("/protected", func(c *gin.Context) {
		c.String(200, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/protected", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Should return 401 without token, got %d", w.Code)
	}
}

func TestAuthRequiredInvalidToken(t *testing.T) {
	r := gin.New()
	r.Use(AuthRequired())
	r.GET("/protected", func(c *gin.Context) {
		c.String(200, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Should return 401 with invalid token, got %d", w.Code)
	}
}

func TestAuthRequiredExpiredToken(t *testing.T) {
	r := gin.New()
	r.Use(AuthRequired())
	r.GET("/protected", func(c *gin.Context) {
		c.String(200, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJleHAiOjF9.invalid")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Should return 401 with expired token, got %d", w.Code)
	}
}

func TestGenerateRequestID(t *testing.T) {
	id := generateRequestID()
	if id == "" {
		t.Error("Request ID should not be empty")
	}
	if len(id) < 10 {
		t.Errorf("Request ID seems too short: %s", id)
	}
}

func TestRandomHex(t *testing.T) {
	hex := randomHex(8)
	if len(hex) != 8 {
		t.Errorf("Expected 8 chars, got %d", len(hex))
	}
	for _, c := range hex {
		if !strings.ContainsRune("0123456789abcdef", c) {
			t.Errorf("Invalid hex character: %c", c)
		}
	}
}
