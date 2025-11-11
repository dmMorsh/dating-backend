package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	data_access "dating-backend/internal/data-access"
	"dating-backend/internal/logging"
	middleware "dating-backend/internal/middleware"
	"dating-backend/internal/models"
	"dating-backend/internal/realtime"

	"github.com/gorilla/schema"
)

var decoder = schema.NewDecoder()

type SwipeRequest struct {
	TargetID int64  `json:"target_id"`
	Action   string `json:"action"` // "like" или "dislike"
}

// SwipeHandler processes a swipe action (like or dislike) from the authenticated user.
// It updates the swipe record in the database and checks for mutual likes to create a match.
// On a mutual like, it creates a chat and sends real-time notifications to both users.
// If no match occurs, it simply acknowledges the swipe action.
// Expected JSON request body:
// {
//     "target_id": <int64>,
//     "action": "like" | "dislike"
// }
func SwipeHandler(w http.ResponseWriter, r *http.Request) {
	userID, authErr := middleware.UserIDFromContext(r.Context())
	if authErr != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req SwipeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logging.Log.Warnf("swipe: decode error: %v", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.TargetID == userID {
		logging.Log.Warnf("swipe: user %d tried to swipe on themselves", userID)
		http.Error(w, "target_id can't be yours", http.StatusBadRequest)
		return
	}

	if req.Action != "like" && req.Action != "dislike" {
		logging.Log.Warnf("swipe: invalid action '%s' from user %d", req.Action, userID)
		http.Error(w, "invalid action", http.StatusBadRequest)
		return
	}

	// Put or update the swipe record
	if err := data_access.UpsertSwipe(userID, req.TargetID, req.Action); err != nil {
		logging.Log.Errorf("swipe: upsert error user=%d target=%d action=%s: %v", userID, req.TargetID, req.Action, err)
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	// Check for mutual likes
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
						// Send real-time notifications to both users
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

// GetMyFollowersHandler retrieves the list of users who have liked the authenticated user.
// It responds with a JSON array of user profiles.
func GetMyFollowersHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		logging.Log.Warnf("get followers: unauthorized: %v", err)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	profiles, err := data_access.GetUserFollowers(userID)
	if err != nil {
		logging.Log.Errorf("get followers: db error user=%d: %v", userID, err)
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profiles)
}

// GetSwipeCandidatesHandler retrieves a list of user profiles that the authenticated user
// has not swiped on yet, applying optional filters from SimpleFilter.
// It responds with a JSON array of user profiles.
// Expected query parameters can include those defined in SimpleFilter.
// For example: ?min_age=18&max_age=30&gender=female
func GetSwipeCandidatesHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		logging.Log.Warnf("get swipe candidates: unauthorized: %v", err)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	filter := models.SimpleFilter{}
	if err := decoder.Decode(&filter, r.URL.Query()); err != nil {
		logging.Log.Warnf("get swipe candidates: decode filter error: %v", err)
		http.Error(w, "invalid query", http.StatusBadRequest)
		return
	}

	profiles, err := data_access.GetSwipeCandidates(userID, &filter)
	if err != nil {
		logging.Log.Errorf("get swipe candidates: db error user=%d: %v", userID, err)
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profiles)
}

// Only for testing purposes
func ClearMySwipesHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		logging.Log.Warnf("clear swipes: unauthorized: %v", err)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	
	err = data_access.ClearSwipesForUser(userID)
	if err != nil {
		logging.Log.Errorf("clear swipes: db error user=%d: %v", userID, err)
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(true)
}