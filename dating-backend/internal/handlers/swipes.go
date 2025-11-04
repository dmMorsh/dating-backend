package handlers

import (
	data_access "dating-backend/internal/data-access"
	middleware "dating-backend/internal/middleware"
	"dating-backend/internal/models"
	"dating-backend/internal/realtime"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/schema"
)

var decoder = schema.NewDecoder()

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

			isNew, chatID, err := data_access.CreateOrGetChat(userID, req.TargetID)
			if isNew{
				var msgMatch = models.Message{
						ChatID:  chatID,
						Content: "It's a match! 🎉",
					}
					_,_ = data_access.SaveMessage(&msgMatch)

				if err == nil {
					// Отправляем уведомления обоим участникам через WebSocket
					msg := map[string]any{
						"type":    "match",
						"content": "It's a match! 🎉",
						"chat_id": chatID,
						"user_id": req.TargetID,
					}
					realtime.ChatHub.SendToUser(userID, msg)
					msg["user_id"] = userID
					realtime.ChatHub.SendToUser(req.TargetID, msg)
				}
			}
			json.NewEncoder(w).Encode(map[string]string{
				"status": "match",
				"content": fmt.Sprintf("It's a match with user %d!", req.TargetID),
			})

			return
		}
	}

	json.NewEncoder(w).Encode(map[string]string{"status": req.Action})
}

func MyFollowersHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	profiles, err := data_access.GetUserFollowers(userID)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profiles)
}

func GetSwipeCandidatesHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	filter := models.SimpleFilter{}
    if err := decoder.Decode(&filter, r.URL.Query()); err != nil {
        http.Error(w, "invalid query", http.StatusBadRequest)
        return
    }

	profiles, err := data_access.GetSwipeCandidates(userID, &filter)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profiles)
}

func ClearMySwipesHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	
	err = data_access.ClearSwipesForUser(userID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(true)
}