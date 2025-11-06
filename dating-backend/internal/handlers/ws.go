package handlers

import (
	"net/http"

	crypto "crypto/rand"
	data_access "dating-backend/internal/data-access"
	models "dating-backend/internal/models"
	"dating-backend/internal/realtime"
	"encoding/hex"
	"encoding/json"
	"time"

	"dating-backend/internal/middleware"

	"github.com/gorilla/websocket"
)

//ЧАТ НЕ РАБОТАЕТ В РЕЖИМЕ РАСШИРЕННЫХ ЛОГОВ
//
//
//

var sessionTokens = make(map[string]int64)

// TODO: Use Redis or another in-memory store for session tokens with expiration.
func init() {
	// Clean up expired tokens every 2 minutes
	go func() {
		for {
			time.Sleep(2 * time.Minute)
			for t, uid := range sessionTokens {
				if uid < 0 { // Store negative IDs for expired tokens.
					delete(sessionTokens, t)
				}
			}
		}
	}()
}

// StartWebSocketSession generates a one-time session token for WebSocket connection.
// The token is mapped to the authenticated userID and stored in memory.
// The token should be included as a query parameter when establishing the WebSocket connection.
// Example usage: /ws/chat?session=<token>
func StartWebSocketSession(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	token := generateSessionToken()
	// TODO: store in Redis with expiration
	sessionTokens[token] = userID

	// Expire the token after 30 seconds
	go func(tok string) {
		time.Sleep(30 * time.Second)
		if _, ok := sessionTokens[tok]; ok {
			sessionTokens[tok] = -1 // Mark as expired
		}
	}(token)

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
// The token is mapped to the userID and stored in memory.
// This token should be included as a query parameter when establishing the WebSocket connection.
// Example usage: /ws/chat?session=<token>
// Note: This function should be called after user authentication.
// Example: token, err := GenerateChatSessionToken(userID)
// The token should be sent to the client to use in the WebSocket connection.
// The token is valid for one-time use only.
func ChatWebSocketHandler(w http.ResponseWriter, r *http.Request) {
	session := r.URL.Query().Get("session")
	userID, ok := sessionTokens[session]
	if !ok || userID < 0 {
		http.Error(w, "invalid session token", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "failed to upgrade", http.StatusInternalServerError)
		return
	}

	realtime.ChatHub.Add(userID, conn)
	delete(sessionTokens, session) // onetime use

	defer func() {
		realtime.ChatHub.Remove(userID)
		conn.Close()
	}()

	for {
		var msg struct {
			ChatID		int64	`json:"chat_id"`
			SenderID	int64	`json:"sender_id"`
			ReceiverID	int64	`json:"receiver_id"`
			Content		string	`json:"content"`
		}
		if err := conn.ReadJSON(&msg); err != nil {
			break // connection closed or error
		}

		// save message to DB
		msgId, saveErr := SaveMessage(msg.ChatID, userID, msg.ReceiverID, msg.Content)
		if saveErr != nil {
			conn.WriteJSON(map[string]string{"error": "failed to save message"})
			continue
		}

		// send to receiver if online
		realtime.ChatHub.SendToUser(msg.ReceiverID, map[string]interface{}{
			"id":		msgId,
			"type":		"message",
			"content":	msg.Content,
			"chat_id":	msg.ChatID,
			"user_id":	userID,
		})
	}
}

func SaveMessage(chatID, senderID, receiverID int64, content string) (int64, error) {
	msg := models.Message{
		ChatID:     chatID,
		SenderID:   senderID,
		ReceiverID: receiverID,
		Content:    content,
	}
	return data_access.SaveMessage(&msg)
}
