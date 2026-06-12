package handler

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/CodingFervor/live-commerce-bi/internal/service"
	"github.com/CodingFervor/live-commerce-bi/pkg/jwt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Validate Origin header against allowed origins
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true // non-browser clients
		}
		// Allow configured origins or localhost in development
		allowed := []string{
			"http://localhost:3000", "http://localhost:8080",
			"http://localhost:5173", "http://127.0.0.1:3000",
		}
		for _, a := range allowed {
			if origin == a {
				return true
			}
		}
		return false
	},
	HandshakeTimeout: 10 * time.Second,
}

// WebSocketHandler handles WebSocket connections
type WebSocketHandler struct {
	hub *service.Hub
}

func NewWebSocketHandler() *WebSocketHandler {
	return &WebSocketHandler{hub: service.GetHub()}
}

// ServeWS upgrades HTTP connection to WebSocket (requires auth token)
func (h *WebSocketHandler) ServeWS(c *gin.Context) {
	// Validate JWT token from query parameter
	tokenStr := c.Query("token")
	if tokenStr == "" {
		// Also accept token from Sec-WebSocket-Protocol header
		if proto := c.GetHeader("Sec-WebSocket-Protocol"); proto != "" {
			parts := strings.SplitN(proto, ",", 2)
			tokenStr = strings.TrimSpace(parts[0])
		}
	}
	if tokenStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	claims, err := jwt.ParseToken(tokenStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	// Store user info in WebSocket connection context
	_ = claims

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[WS] Upgrade failed: %v", err)
		return
	}

	roomID := c.Query("room_id")
	client := &service.Client{
		Hub:    h.hub,
		Send:   make(chan []byte, 256),
		RoomID: roomID,
	}

	h.hub.Register <- client

	go h.writePump(client, conn)
	go h.readPump(client, conn)
}

func (h *WebSocketHandler) writePump(client *service.Client, conn *websocket.Conn) {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (h *WebSocketHandler) readPump(client *service.Client, conn *websocket.Conn) {
	defer func() {
		h.hub.Unregister <- client
		conn.Close()
	}()

	conn.SetReadLimit(4096)
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("[WS] Error: %v", err)
			}
			break
		}
		// Client messages can trigger actions
		_ = message
	}
}

// WSStats returns WebSocket connection statistics
func (h *WebSocketHandler) WSStats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"total_clients": h.hub.ClientCount(),
		"total_rooms":   h.hub.RoomCount(),
	})
}

// SubscribeRoom subscribes a WebSocket client to a specific room
func (h *WebSocketHandler) SubscribeRoom(c *gin.Context) {
	roomID := strconv.FormatInt(0, 10)
	if id, err := strconv.ParseInt(c.Param("room_id"), 10, 64); err == nil {
		roomID = strconv.FormatInt(id, 10)
	}
	// Redirect to WS with room_id query param
	c.Redirect(http.StatusTemporaryRedirect, "/ws?room_id="+roomID)
}
