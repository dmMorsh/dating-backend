package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	data_access "dating-backend/internal/data-access"
	middleware "dating-backend/internal/middleware"
	models "dating-backend/internal/models"
)

func SendMessageHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var msg models.Message
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	if msg.ReceiverID == 0 || msg.Content == "" {
		http.Error(w, "missing fields", http.StatusBadRequest)
		return
	}
	msg.SenderID = userID

	chatId, err := data_access.CreateOrGetChat(userID, msg.ReceiverID)
	if err != nil {
		http.Error(w, "failed to get chat", http.StatusInternalServerError)
		return
	}
	msg.ChatID = chatId

	if err := data_access.SaveMessage(&msg); err != nil {
		http.Error(w, "failed to save", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(msg)
}

func GetMessagesHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	partnerIDStr := r.URL.Query().Get("partner_id")
	if partnerIDStr == "" {
		http.Error(w, "missing partner_id", http.StatusBadRequest)
		return
	}
	partnerID, err := strconv.ParseInt(partnerIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid partner_id", http.StatusBadRequest)
		return
	}

	msgs, err := data_access.GetMessagesBetweenUsers(userID, partnerID)
	if err != nil {
		http.Error(w, "failed to fetch messages", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(msgs)
}