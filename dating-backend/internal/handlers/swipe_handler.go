package handlers

import (
	data_access "dating-backend/internal/data-access"
	middleware "dating-backend/internal/middleware"
	"dating-backend/internal/realtime"
	"encoding/json"
	"fmt"
	"net/http"
)

type SwipeRequest struct {
	TargetID int64  `json:"target_id"`
	Action   string `json:"action"` // "like" или "dislike"
}

func SwipeHandler(w http.ResponseWriter, r *http.Request) {
	userID, authErr := middleware.UserIDFromContext(r.Context())
	if authErr != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req SwipeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.TargetID == userID {
		http.Error(w, "target_id can't be yours", http.StatusBadRequest)
		return
	}

	if req.Action != "like" && req.Action != "dislike" {
		http.Error(w, "invalid action", http.StatusBadRequest)
		return
	}

	// Вставляем или обновляем свайп
	if err := data_access.UpsertSwipe(userID, req.TargetID, req.Action); err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	// Проверяем взаимный лайк
	if req.Action == "like" {
		mutual, err := data_access.HasLiked(req.TargetID, userID)

		if err == nil && mutual {

			chatID, err := data_access.CreateOrGetChat(userID, req.TargetID)
			if err == nil {
				// Отправляем уведомления обоим участникам через WebSocket
				msg := map[string]any{
					"type":    "match",
					"message": "It's a match! 🎉",
					"chat_id": chatID,
					"user_id": req.TargetID,
				}
				realtime.ChatHub.SendToUser(userID, msg)
				msg["user_id"] = userID
				realtime.ChatHub.SendToUser(req.TargetID, msg)
			}
			
			json.NewEncoder(w).Encode(map[string]string{
				"status": "match",
				"message": fmt.Sprintf("It's a match with user %d!", req.TargetID),
			})

			return
		}
	}

	json.NewEncoder(w).Encode(map[string]string{"status": req.Action})
}
