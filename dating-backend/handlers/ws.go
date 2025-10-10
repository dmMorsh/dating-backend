package handlers

import (
	"net/http"

	data_access "dating-backend/data-access"
	middleware "dating-backend/middleware"
	models "dating-backend/models"
	"dating-backend/realtime"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Разрешаем любые домены (пока)
	},
}

func ChatWebSocketHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "failed to upgrade", http.StatusInternalServerError)
		return
	}

	realtime.ChatHub.Add(userID, conn)

	// После добавления пользователя в хаб, отправляем ему непрочитанные сообщения
	unread, err := data_access.GetUnreadMessages(userID)
	if err == nil && len(unread) > 0 {
		conn.WriteJSON(map[string]interface{}{
			"type":      "unread_messages",
			"messages":  unread,
		})
		data_access.MarkMessagesAsRead(userID)
	}

	defer func() {
		realtime.ChatHub.Remove(userID)
		conn.Close()
	}()

	for {
		var msg struct {
			ReceiverID int64  `json:"receiver_id"`
			Content    string `json:"content"`
		}
		if err := conn.ReadJSON(&msg); err != nil {
			break // соединение закрыто
		}

		// сохраняем в БД
		saveErr := SaveMessage(userID, msg.ReceiverID, msg.Content)
		if saveErr != nil {
			conn.WriteJSON(map[string]string{"error": "failed to save message"})
			continue
		}

		// отправляем получателю, если он онлайн
		realtime.ChatHub.SendToUser(msg.ReceiverID, map[string]interface{}{
			"sender_id": userID,
			"content":   msg.Content,
		})
	}
}

func SaveMessage(senderID, receiverID int64, content string) error {
	msg := models.Message{
		SenderID:   senderID,
		ReceiverID: receiverID,
		Content:    content,
	}
	return data_access.SaveMessage(&msg)
}
