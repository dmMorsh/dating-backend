package realtime

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	pingInterval   = 30 * time.Second
	writeWait      = 5 * time.Second
)

func StartPingLoop() {
	go func() {
		ticker := time.NewTicker(pingInterval)
		defer ticker.Stop()

		for range ticker.C {
			ChatHub.mu.RLock()
			clients := make([]*Connection, 0, len(ChatHub.clients))
			for _, c := range ChatHub.clients {
				clients = append(clients, c)
			}
			ChatHub.mu.RUnlock()

			for _, c := range clients {
				c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					log.Printf("ws: ping failed for user=%d: %v. Removing client.", c.UserID, err)
					ChatHub.Remove(c.UserID)
					c.Conn.Close()
				}
			}
		}
	}()
}
