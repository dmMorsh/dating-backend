package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	data_access "dating-backend/internal/data-access"
)

type ctxKey string

const userIDKey ctxKey = "userID"

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "missing or invalid authorization header", http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var userID int64
		var exp time.Time
		err := data_access.DB.QueryRow(`SELECT user_id, access_expires FROM sessions WHERE access_token=?`, token).Scan(&userID, &exp)
		if err != nil || time.Now().After(exp) {
			http.Error(w, "Token expired or invalid", http.StatusUnauthorized)
			return
		}

		// Передаем userID в контекст
		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next(w, r.WithContext(ctx))
	}
}