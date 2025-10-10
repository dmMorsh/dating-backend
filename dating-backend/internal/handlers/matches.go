package handlers

import (
	"encoding/json"
	"net/http"

	data_access "dating-backend/internal/data-access"
	middleware "dating-backend/internal/middleware"
)

func MatchesHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	matches, err := data_access.GetMatches(userID)
	if err != nil {
		http.Error(w, "failed to load matches", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(matches)
}