package ws

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow all origins in dev; restrict in production.
	CheckOrigin: func(r *http.Request) bool { return true },
}

type chatHandler struct {
	hub *Hub
}

func NewChatHandler(hub *Hub) *chatHandler {
	return &chatHandler{hub: hub}
}

// ServeWS upgrades the HTTP connection to WebSocket and registers the client.
func (h *chatHandler) ServeWS(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		slog.Error("ws upgrade failed", "err", err)
		return
	}

	client := &Client{
		hub:  h.hub,
		send: make(chan []byte, 256),
		conn: conn,
	}

	h.hub.register <- client

	go client.writePump(conn)
	client.readPump(conn) // blocks until connection closes
}

// readPump reads messages from the WebSocket and broadcasts them.
func (c *Client) readPump(conn *websocket.Conn) {
	defer func() {
		c.hub.unregister <- c
		conn.Close()
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Warn("ws: unexpected close", "err", err)
			}
			break
		}
		c.hub.Broadcast(msg)
	}
}

// writePump drains the client's send channel to the WebSocket.
func (c *Client) writePump(conn *websocket.Conn) {
	defer conn.Close()

	for msg := range c.send {
		if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			slog.Warn("ws: write error", "err", err)
			return
		}
	}
}
