package handlers

import (
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	crypto "crypto/rand"
	"dating-backend/internal/realtime"

	"dating-backend/internal/logging"
	"dating-backend/internal/middleware"

	"github.com/gorilla/websocket"
)

// Session tokens are handled via the SessionStore abstraction in
// `internal/realtime/session_store.go`. By default it uses an in-memory
// store but can be replaced with a Redis-backed store by assigning
// `realtime.DefaultSessionStore = realtime.NewRedisSessionStore(...)` from
// `main` before the server starts.

// StartWebSocketSession generates a one-time session token for WebSocket connection.
// The token is mapped to the authenticated userID and stored.
func StartWebSocketSession(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		logging.Log.Warnf("ws/start: unauthorized: %v", err)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	token := generateSessionToken()
	// store token with TTL; the default store uses in-memory map with cleaner,
	// but this can be swapped to Redis by assigning realtime.DefaultSessionStore
	// before the server starts.
	if err := realtime.DefaultSessionStore.Set(token, userID, 30*time.Second); err != nil {
		logging.Log.Errorf("ws/start: failed to store session token: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"session_token": token,
	})
}

func generateSessionToken() string {
	b := make([]byte, 32)
	if _, err := crypto.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow lack of origin or specific origin
        origin := r.Header.Get("Origin")
        return origin == "" || origin == "https://intellyjourney.ru"
	},
}

// GenerateChatSessionToken creates a one-time session token for WebSocket chat connection.
// The token is mapped to the userID and stored.
// This token should be included as a query parameter when establishing the WebSocket connection.
// Example usage: /ws/chat?session=<token>
// Note: This function should be called after user authentication.
// The token is valid for one-time use only.
func ChatWebSocketHandler(w http.ResponseWriter, r *http.Request) {
	session := r.URL.Query().Get("session")
	userID, ok, err := realtime.DefaultSessionStore.Get(session)
	if err != nil {
		logging.Log.Errorf("ws: session lookup error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if !ok {
		logging.Log.Warnf("ws: invalid session token: %s", session)
		http.Error(w, "invalid session token", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logging.Log.Errorf("ws: upgrade error: %v", err)
		http.Error(w, "failed to upgrade", http.StatusInternalServerError)
		return
	}

	realtime.ChatHub.Add(userID, conn)
	
	// one-time use - remove token from the store
	if err := realtime.DefaultSessionStore.Delete(session); err != nil {
		logging.Log.Errorf("ws: failed to delete session token: %v", err)
	}

	defer func() {
		realtime.ChatHub.Remove(userID)
		conn.Close()
	}()

	// Ping/Pong to keep the connection alive 
	conn.SetReadLimit(512) // TODO: adjust limit as needed 
	conn.SetReadDeadline(time.Now().Add(60 * time.Second)) 
	conn.SetPongHandler(func(string) error { 
		conn.SetReadDeadline(time.Now().Add(60 * time.Second)) 
		return nil 
	})

	for {
		var msg struct {
            Type       string  `json:"type"`
            ChatID     int64   `json:"chat_id,omitempty"`
            ReceiverID int64   `json:"receiver_id,omitempty"`
            Content    string  `json:"content,omitempty"`
            MessageID  int64   `json:"message_id,omitempty"`
        }

		if err := conn.ReadJSON(&msg); err != nil {
			if websocket.IsCloseError(err, websocket.CloseAbnormalClosure) {
				logging.Log.Warnf("user=%d disconnected unexpectedly: %v", userID, err)
				break
			}
			logging.Log.Errorf("ws: read json error user=%d: %v", userID, err)
				break
		}

		switch msg.Type {

        case "typing":
            realtime.ChatHub.SendToUser(msg.ReceiverID, map[string]interface{}{
                "type":		"typing",
                "chat_id":	msg.ChatID,
                "user_id":	userID,
            })
			
        case "delivered":
			realtime.ChatHub.SendToUser(msg.ReceiverID, map[string]interface{}{
				"type":       "delivered",
				"chat_id":    msg.ChatID,
				"user_id":    userID,
				"message_id": msg.MessageID,
			})
        }
    }
}
