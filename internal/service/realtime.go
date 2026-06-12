package service

import (
	"context"
	"sync"
	"time"

	"github.com/CodingFervor/live-commerce-bi/internal/cache"
)

// ═══ WebSocket Real-time Push Engine ═══

type Client struct {
	Hub    *Hub
	Send   chan []byte
	RoomID string
}

type Hub struct {
	clients    map[*Client]bool
	rooms      map[string]map[*Client]bool
	broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
	mu         sync.RWMutex
}

var defaultHub *Hub

func NewHub() *Hub {
	h := &Hub{
		clients:    make(map[*Client]bool),
		rooms:      make(map[string]map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
	defaultHub = h
	return h
}

func GetHub() *Hub {
	if defaultHub == nil {
		defaultHub = NewHub()
	}
	return defaultHub
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.clients[client] = true
			if client.RoomID != "" {
				if h.rooms[client.RoomID] == nil {
					h.rooms[client.RoomID] = make(map[*Client]bool)
				}
				h.rooms[client.RoomID][client] = true
			}
			h.mu.Unlock()

		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
				if client.RoomID != "" {
					if room, ok := h.rooms[client.RoomID]; ok {
						delete(room, client)
						if len(room) == 0 {
							delete(h.rooms, client.RoomID)
						}
					}
				}
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) BroadcastToAll(message []byte) {
	h.broadcast <- message
}

func (h *Hub) BroadcastToRoom(roomID string, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if room, ok := h.rooms[roomID]; ok {
		for client := range room {
			select {
			case client.Send <- message:
			default:
				close(client.Send)
				delete(h.clients, client)
			}
		}
	}
}

func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

func (h *Hub) RoomCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.rooms)
}

// RedisSubscriber listens to Redis Pub/Sub and forwards to WebSocket hub
func RedisSubscriber(ctx context.Context) {
	rdb := cache.Get()
	if rdb == nil {
		return
	}
	sub := rdb.Subscribe(ctx, "live:metrics:update", "live:room:alert", "live:room:gmv")
	defer sub.Close()

	ch := sub.Channel()
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-ch:
			hub := GetHub()
			switch msg.Channel {
			case "live:metrics:update":
				hub.BroadcastToAll([]byte(msg.Payload))
			case "live:room:alert", "live:room:gmv":
				hub.BroadcastToAll([]byte(msg.Payload))
			}
		}
	}
}

// PushRealtimeMetric pushes a metric update through Redis -> WebSocket pipeline
func PushRealtimeMetric(ctx context.Context, channel string, data []byte) error {
	rdb := cache.Get()
	if rdb == nil {
		return nil
	}
	cacheKey := "realtime:" + channel
	_ = cache.SetJSON(ctx, cacheKey, string(data), 5*time.Minute)
	return rdb.Publish(ctx, channel, string(data)).Err()
}
