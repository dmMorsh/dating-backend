package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	data_access "dating-backend/internal/data-access"
	"dating-backend/internal/logging"
	middleware "dating-backend/internal/middleware"
	models "dating-backend/internal/models"
	"dating-backend/internal/realtime"
)

// SendMessageHandler handles sending a message from the authenticated user
// to another user. It ensures a chat exists between the users, saves the
// message, and notifies the receiver via WebSocket if connected.
// Expects a JSON body with filled "receiver_id" and "content" fields in Message model.
// Example request body:
// {
//     "receiver_id": 42,
//     "content": "Hello there!",
//     "chat_id": 0,
//     "sender_id": 0
// }
func SendMessageHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		logging.Log.Warnf("send message: unauthorized: %v", err)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var msg models.Message
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		logging.Log.Warnf("send message: decode error: %v", err)
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	if msg.ChatID == 0 || msg.ReceiverID == 0 || msg.Content == "" {
		logging.Log.Warnf("send message: missing fields from user=%d receiver=%d chat=%d", userID, msg.ReceiverID, msg.ChatID)
		http.Error(w, "missing fields", http.StatusBadRequest)
		return
	}
	msg.SenderID = userID

	var msgId int64
	if msgId, err = data_access.SaveMessage(&msg); err != nil {
		logging.Log.Errorf("send message: save error chat=%d sender=%d receiver=%d: %v", msg.ChatID, msg.SenderID, msg.ReceiverID, err)
		http.Error(w, "failed to save", http.StatusInternalServerError)
		return
	}
	msg.ID = msgId
	msg.CreatedAt = time.Now()

	realtime.ChatHub.SendToUser(msg.ReceiverID, map[string]interface{}{
		"type":		"message",
		"id":		msgId,
		"chat_id":	msg.ChatID,
		"user_id":	userID,
		"content":	msg.Content,
	})

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(msg)
}

// GetChatsHandler retrieves all chats for the authenticated user.
// Each chat includes the latest message and unread count.
func GetChatsHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	msgs, err := data_access.GetChatsForUser(userID)
	if err != nil {
		logging.Log.Errorf("get chats: db error user=%d: %v", userID, err)
		http.Error(w, "failed to fetch chats", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(msgs)
}

// GetChatMessagesHandler retrieves messages for a specific chat.
// The chat ID is taken from the URL path.
// Supports optional query parameters:
// - limit: maximum number of messages to return (default 50, max 200)
// - before_id: fetch messages with IDs less than this value
// - after_id: fetch messages with IDs greater than this value
// Example: GET /chat/messages/{chatId}?limit=100&before_id=500
func GetChatMessagesHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/chat/messages/")
	chatId, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		logging.Log.Warnf("get chat messages: invalid chat id '%s': %v", idStr, err)
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	beforeStr := r.URL.Query().Get("before_id")
	afterStr := r.URL.Query().Get("after_id")

	limit := 50
	if limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 && parsed <= 200 {
			limit = parsed
		}
	}

	var beforeID *int64
	if beforeStr != "" {
		if parsed, err := strconv.ParseInt(beforeStr, 10, 64); err == nil {
			beforeID = &parsed
		}
	}

	var afterID *int64
	if afterStr != "" {
		if parsed, err := strconv.ParseInt(afterStr, 10, 64); err == nil {
			afterID = &parsed
		}
	}

	msgs, err := data_access.GetMessagesForChat(chatId, beforeID, afterID, limit)
	if err != nil {
		logging.Log.Errorf("get chat messages: db error chat=%d: %v", chatId, err)
		http.Error(w, "failed to fetch messages", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(msgs)
}

// MarkChatMessagesAsReadHandler marks all messages in a chat as read for the
// authenticated user.
// The chat ID is taken from the URL path.
// Example: POST /chat/read
func MarkChatMessagesAsReadHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		logging.Log.Warnf("mark chat read: unauthorized: %v", err)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		ChatId int64 `json:"chat_id"`
		ReceiverID int64 `json:"receiver_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logging.Log.Warnf("mark chat read: decode error: %v", err)
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	res, err := data_access.MarkMessagesAsReadForChat(req.ChatId, userID)
	if err != nil {
		logging.Log.Errorf("mark chat read: db error chat=%d user=%d: %v", req.ChatId, userID, err)
		http.Error(w, "failed to set messages", http.StatusInternalServerError)
		return
	}

	realtime.ChatHub.SendToUser(req.ReceiverID, map[string]interface{}{
		"type":       "read_chat",
		"chat_id":    req.ChatId,
		"user_id":    userID,
	})

	json.NewEncoder(w).Encode(res)
}

type MarkMsgReadRequest struct {
	MessageIDs []int64 `json:"message_ids"`
	ChatId    int64    `json:"chat_id"`
	ReceiverID int64   `json:"receiver_id"`
}

// MarkMessagesReadHandler marks messages as read based on provided message IDs.
// Expects a JSON body with a "message_ids" field containing an array of int64 IDs.
func MarkMessagesReadHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		logging.Log.Warnf("mark chat read: unauthorized: %v", err)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var req MarkMsgReadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logging.Log.Warnf("mark messages read: decode error: %v", err)
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if len(req.MessageIDs) == 0 {
		logging.Log.Warn("mark messages read: empty message id list")
		http.Error(w, "no message ids provided", http.StatusBadRequest)
		return
	}

	_, err = data_access.MarkMessagesAsRead(req.MessageIDs)
	if err != nil {
		logging.Log.Errorf("mark messages read: db error: %v", err)
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	realtime.ChatHub.SendToUser(req.ReceiverID, map[string]interface{}{
		"type":       "read_messages",
		"chat_id":    req.ChatId,
		"user_id":    userID,
		"message_id": req.MessageIDs,
	})

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}
