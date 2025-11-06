package middleware

import (
	"context"
	data_access "dating-backend/internal/data-access"
	"net/http"
	"strings"
	"time"
)

type ctxKey string

const userIDKey ctxKey = "userID"

// AuthMiddleware validates Bearer token from the Authorization header.
//
// On success it injects the user id into the request context (use
// `UserIDFromContext` to retrieve it) and calls the next handler. On
// failure it writes an HTTP 401 response and does not call next.
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

		// Inject userID into context
		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next(w, r.WithContext(ctx))
	}
}