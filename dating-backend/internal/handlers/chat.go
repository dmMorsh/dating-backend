package handlers

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"time"

	"dating-backend/internal/middleware"
)

var sessionTokens = make(map[string]int64)

func init() {
	// Чистим просроченные токены каждые 2 минуты
	go func() {
		for {
			time.Sleep(2 * time.Minute)
			//now := time.Now()
			for t, uid := range sessionTokens {
				if uid < 0 { // мы храним отрицательные id для просроченных токенов
					delete(sessionTokens, t)
				}
			}
		}
	}()
}

func StartWebSocketSession(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	token := generateSessionToken()
	sessionTokens[token] = userID

	// Токен живёт 30 секунд
	go func(tok string) {
		time.Sleep(30 * time.Second)
		if _, ok := sessionTokens[tok]; ok {
			sessionTokens[tok] = -1 // помечаем как просроченный
		}
	}(token)

	json.NewEncoder(w).Encode(map[string]string{
		"session_token": token,
	})
}

func generateSessionToken() string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, 32)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
