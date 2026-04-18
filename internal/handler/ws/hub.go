package ws

import (
	"log/slog"
	"sync"
)

// Client represents a single WebSocket connection.
type Client struct {
	hub  *Hub
	send chan []byte
	conn interface{ Close() error } // set by chatHandler after upgrade
}

// Hub manages all active WebSocket connections and broadcasts messages.
type Hub struct {
	mu         sync.RWMutex
	clients    map[*Client]struct{}
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]struct{}),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run starts the hub event loop. Call in a goroutine.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = struct{}{}
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()

		case msg := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- msg:
				default:
					// slow client — drop and disconnect
					slog.Warn("ws: slow client, dropping", "client", client)
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast enqueues a message to all connected clients.
func (h *Hub) Broadcast(msg []byte) {
	h.broadcast <- msg
}
