package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	data_access "dating-backend/internal/data-access"
	middleware "dating-backend/internal/middleware"
)

func RecommendationsHandler(w http.ResponseWriter, r *http.Request) {
    userID, err := middleware.UserIDFromContext(r.Context())
    if err != nil {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return
    }

    // Optional params: limit and maxDistance
    q := r.URL.Query()
    limit := 20
    if s := q.Get("limit"); s != "" {
        if v, err2 := strconv.Atoi(s); err2 == nil && v > 0 {
            limit = v
        }
    }
    maxDist := 50.0 // km on default
    if s := q.Get("max_distance"); s != "" {
        if v, err2 := strconv.ParseFloat(s, 64); err2 == nil && v > 0 {
            maxDist = v
        }
    }

    recs, err := data_access.GetRecommendations(userID, limit, maxDist)
    if err != nil {
        http.Error(w, "failed to get recommendations", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(recs)
}
