package handlers

import (
	data_access "dating-backend/data-access"
	middleware "dating-backend/middleware"
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

	if req.Action != "like" && req.Action != "dislike" {
		http.Error(w, "invalid action", http.StatusBadRequest)
		return
	}

	// Вставляем или обновляем свайп
	_, err := data_access.DB.Exec(`
		INSERT OR REPLACE INTO swipes (user_id, target_id, action)
		VALUES (?, ?, ?)
	`, userID, req.TargetID, req.Action)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	// Проверяем взаимный лайк
	if req.Action == "like" {
		var count int
		err = data_access.DB.QueryRow(`
			SELECT COUNT(*) FROM swipes
			WHERE user_id = ? AND target_id = ? AND action = 'like'
		`, req.TargetID, userID).Scan(&count)

		if err == nil && count > 0 {
			// Взаимный лайк → создаём мэтч
			_, _ = data_access.DB.Exec(`
				INSERT OR IGNORE INTO matches (user1_id, user2_id)
				VALUES (?, ?)
			`, userID, req.TargetID)

			// И создаём чат
			_, _ = data_access.DB.Exec(`
				INSERT OR IGNORE INTO chats (user1_id, user2_id)
				VALUES (?, ?)
			`, userID, req.TargetID)

			json.NewEncoder(w).Encode(map[string]string{
				"status": "match",
				"message": fmt.Sprintf("It's a match with user %d!", req.TargetID),
			})
			return
		}
	}

	json.NewEncoder(w).Encode(map[string]string{"status": req.Action})
}
