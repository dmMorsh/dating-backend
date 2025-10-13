package handlers

import (
	"net/http"

	data_access "dating-backend/internal/data-access"
	models "dating-backend/internal/models"
	"dating-backend/internal/realtime"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Разрешаем любые домены (пока)
	},
}

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
	delete(sessionTokens, session) // одноразовый

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
			break // соединение закрыто
		}

		// сохраняем в БД
		saveErr := SaveMessage(msg.ChatID, userID, msg.ReceiverID, msg.Content)
		if saveErr != nil {
			conn.WriteJSON(map[string]string{"error": "failed to save message"})
			continue
		}

		// отправляем получателю, если он онлайн
		realtime.ChatHub.SendToUser(msg.ReceiverID, map[string]interface{}{
			"type":		"message",
			"content":	msg.Content,
			"chat_id":	msg.ChatID,
			"user_id":	userID,
		})
	}
}

func SaveMessage(chatID, senderID, receiverID int64, content string) error {
	msg := models.Message{
		ChatID:     chatID,
		SenderID:   senderID,
		ReceiverID: receiverID,
		Content:    content,
	}
	return data_access.SaveMessage(&msg)
}
