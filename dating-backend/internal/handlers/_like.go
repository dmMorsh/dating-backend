package handlers

// import (
// 	"encoding/json"
// 	"net/http"
// 	"strconv"
// 	"strings"

// 	data_access "dating-backend/internal/data-access"
// 	middleware "dating-backend/internal/middleware"
// 	"dating-backend/internal/realtime"
// )

// func LikeHandler(w http.ResponseWriter, r *http.Request) {
// 	userID, err := middleware.UserIDFromContext(r.Context())
// 	if err != nil {
// 		http.Error(w, "unauthorized", http.StatusUnauthorized)
// 		return
// 	}
// 	parts := strings.Split(r.URL.Path, "/")
// 	if len(parts) < 3 {
// 		http.Error(w, "invalid request", http.StatusBadRequest)
// 		return
// 	}
// 	targetID, err := strconv.ParseInt(parts[2], 10, 64)
// 	if err != nil {
// 		http.Error(w, "invalid target id", http.StatusBadRequest)
// 		return
// 	}

// 	if err := data_access.AddLike(userID, targetID); err != nil {
// 		http.Error(w, "failed to add like", http.StatusInternalServerError)
// 		return
// 	}

// 	isMatch, err := data_access.IsMatch(userID, targetID)
// 	if err != nil {
// 		http.Error(w, "failed to check match", http.StatusInternalServerError)
// 		return
// 	}

// 	resp := map[string]any{
// 		"match": isMatch,
// 	}
// 	if isMatch {
// 		resp["message"] = "It's a match! 🎉"

// 		// ✅ Создаем чат между пользователями (если еще не создан)
// 		chatID, err := data_access.CreateOrGetChat(userID, targetID)
// 		if err == nil {
// 			// Отправляем уведомления обоим участникам через WebSocket
// 			msg := map[string]any{
// 				"type":    "match",
// 				"message": "It's a match! 🎉",
// 				"chat_id": chatID,
// 				"user_id": targetID,
// 			}
// 			realtime.ChatHub.SendToUser(userID, msg)
// 			msg["user_id"] = userID
// 			realtime.ChatHub.SendToUser(targetID, msg)
// 		}
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(resp)
// }