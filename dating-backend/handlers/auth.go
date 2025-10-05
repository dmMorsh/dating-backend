package handlers

import (
	data_access "dating-backend/data-access"
	models "dating-backend/models"
	utils "dating-backend/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)


func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var newUser models.User
	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Проверка обязательных полей
	if strings.TrimSpace(newUser.Username) == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(newUser.Password) == "" {
		http.Error(w, "Password is required", http.StatusBadRequest)
		return
	}

	// Хэширование пароля
	HashPassword, err := utils.HashPassword(newUser.Password)
	if err != nil {
		http.Error(w, "Hashing password error", http.StatusInternalServerError)
		return
	}
	newUser.Password = HashPassword

	// Вставка пользователя в БД
	stmt, err := data_access.DB.Prepare("INSERT INTO users(username, password, bio, photo_url) VALUES(?,?,?,?)")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = stmt.Exec(newUser.Username, newUser.Password, newUser.Bio, newUser.PhotoURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Регистрация успешна"})
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var credentials struct {
		Username  string `json:"username"`
		Password  string `json:"password"`
		DeviceID  string `json:"device_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	row := data_access.DB.QueryRow("SELECT id, password FROM users WHERE username=?", credentials.Username)
	var id int64
	var storedPassword string
	err := row.Scan(&id, &storedPassword)
	if err != nil {
		http.Error(w, "Invalid username", http.StatusUnauthorized)
		return
	}

	if !utils.CheckPasswordHash(credentials.Password, storedPassword) {
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		return
	}

	row.Scan()

	// Генерация токенов
	accessToken := utils.GenerateToken(32)
	refreshToken := utils.GenerateToken(64)
	accessExp := time.Now().Add(15 * time.Minute)
	refreshExp := time.Now().Add(7 * 24 * time.Hour)

	// Сохраняем в БД
	_, err = data_access.DB.Exec(`INSERT INTO sessions (user_id, device_id, access_token, refresh_token, access_expires, refresh_expires)
	                  VALUES (?, ?, ?, ?, ?, ?)`,
		id, credentials.DeviceID, accessToken, refreshToken, accessExp, refreshExp)
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	resp := map[string]interface{}{
		"message": "Успешный вход",
		"user_id": fmt.Sprint(id),
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"access_expires": accessExp,
	}
	json.NewEncoder(w).Encode(resp)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	_, _ = data_access.DB.Exec(`DELETE FROM sessions WHERE access_token=? OR refresh_token=?`, token, token)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"logged out"}`))
}

func RefreshHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID	  		int64  `json:"user_id"`
		RefreshToken 	string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Проверяем валидность refresh токена
	var refreshExp time.Time
	err := data_access.DB.QueryRow(`SELECT refresh_expires FROM sessions WHERE user_id = ? AND refresh_token = ?`,
		req.UserID, req.RefreshToken).Scan(&refreshExp)
	if err != nil || time.Now().After(refreshExp) {
		http.Error(w, "Invalid or expired refresh token", http.StatusUnauthorized)
		return
	}

	newAccess := utils.GenerateToken(32)
	newExp := time.Now().Add(15 * time.Minute)

	_, err = data_access.DB.Exec(`UPDATE sessions SET access_token=?, access_expires=? WHERE user_id = ? AND refresh_token=?`,
		newAccess, newExp, req.UserID, req.RefreshToken)
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"access_token":  newAccess,
		"access_expires": newExp,
	}
	json.NewEncoder(w).Encode(resp)
}