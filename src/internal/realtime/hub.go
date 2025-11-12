package realtime

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Connection struct {
	UserID int64
	Conn   *websocket.Conn
}

type Hub struct {
	clients map[int64]*Connection
	mu      sync.RWMutex
}

var ChatHub = &Hub{
	clients: make(map[int64]*Connection),
}

func (h *Hub) Add(userID int64, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[userID] = &Connection{UserID: userID, Conn: conn}
}

func (h *Hub) Remove(userID int64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if c, ok := h.clients[userID]; ok {
		c.Conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		c.Conn.Close()
		delete(h.clients, userID)
	}
}

func (h *Hub) SendToUser(userID int64, data interface{}) error {
	h.mu.RLock()
	conn, ok := h.clients[userID]
	h.mu.RUnlock()
	if !ok {
		return nil // user offline
	}
	return conn.Conn.WriteJSON(data)
}