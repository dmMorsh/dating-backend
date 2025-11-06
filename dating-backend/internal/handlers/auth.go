package handlers

import (
	data_access "dating-backend/internal/data-access"
	models "dating-backend/internal/models"
	utils "dating-backend/internal/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// RegisterHandler handles user registration requests.
// It expects a JSON body with username, password, bio, and photo_url fields.
// On success, it responds with a success message. On failure, it responds with an error.
// Method: POST
// Endpoint: /register
// Example request body:
// {
//   "username": "johndoe",
//   "password": "securepassword",
//   "bio": "Hello, I'm John!",
//   "photo_url": "http://example.com/photo.jpg"
// }
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

	// Check required fields
	if strings.TrimSpace(newUser.Username) == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(newUser.Password) == "" {
		http.Error(w, "Password is required", http.StatusBadRequest)
		return
	}

	// Hash password
	HashPassword, err := utils.HashPassword(newUser.Password)
	if err != nil {
		http.Error(w, "Hashing password error", http.StatusInternalServerError)
		return
	}
	newUser.Password = HashPassword

	// Insert into DB new user
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

// LoginHandler handles user login requests.
// It expects a JSON body with username, password, and device_id fields.
// On success, it responds with access and refresh tokens. On failure, it responds with an error.
// Method: POST
// Endpoint: /login
// Example request body:
// {
//   "username": "johndoe",
//   "password": "securepassword",
//   "device_id": "device123"
// }
// Example response body:
// {
//   "user_id": "123",
//   "access_token": "access_token_value",
//   "refresh_token": "refresh_token_value",
//   "access_expires": "2024-01-01T12:00:00Z"
// }
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

	// Generate tokens
	accessToken := utils.GenerateToken(32)
	refreshToken := utils.GenerateToken(64)
	accessExp := time.Now().Add(15 * time.Minute)
	refreshExp := time.Now().Add(30 * 24 * time.Hour)

	// Store tokens in DB
	_, err = data_access.DB.Exec(`INSERT OR REPLACE INTO sessions (user_id, device_id, access_token, refresh_token, access_expires, refresh_expires)
	                  VALUES (?, ?, ?, ?, ?, ?)`,
		id, credentials.DeviceID, accessToken, refreshToken, accessExp, refreshExp)
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	resp := map[string]interface{}{
		"user_id": fmt.Sprint(id),
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"access_expires": accessExp,
	}
	json.NewEncoder(w).Encode(resp)
}

// LogoutHandler handles user logout requests.
// It expects the access token in the Authorization header.
// On success, it responds with a logout confirmation. On failure, it responds with an error.
// Method: POST
// Example request body:
// {
//   "user_id": 123,
//   "device_id": "device123"
// }
// Endpoint: /logout
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var logOutCred struct {
		UserID		int64  `json:"user_id"`
		DeviceID	string `json:"device_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&logOutCred); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	res, err := data_access.DB.Exec(`DELETE FROM sessions WHERE user_id=? AND device_id=? AND access_token=?`,
								logOutCred.DeviceID, logOutCred.DeviceID, token)
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "No active session found", http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"logged out"}`))
}

// RefreshHandler handles token refresh requests.
// It expects a JSON body with user_id and refresh_token fields.
// On success, it responds with a new access token. On failure, it responds with an error.
// Method: POST
// Endpoint: /refresh
// Example request body:
// {
//   "user_id": 123,
//   "refresh_token": "existing_refresh_token"
// }
// Example response body:
// {
//   "access_token": "new_access_token",
//   "access_expires": "2024-01-01T12:00:00Z"
// }
func RefreshHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID	  		int64  `json:"user_id"`
		RefreshToken 	string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Validate refresh token
	var refreshExp time.Time
	err := data_access.DB.QueryRow(`SELECT refresh_expires FROM sessions WHERE user_id = ? AND refresh_token = ?`,
		req.UserID, req.RefreshToken).Scan(&refreshExp)
	if err != nil || time.Now().After(refreshExp) {
		http.Error(w, "Invalid or expired refresh token", http.StatusUnauthorized)
		return
	}

	newAccess := utils.GenerateToken(32)
	newExp := time.Now().Add(15 * time.Minute)
	newRefreshExp := time.Now().Add(30 * 24 * time.Hour)

	_, err = data_access.DB.Exec(`UPDATE sessions SET access_token=?, access_expires=?, refresh_expires=? WHERE user_id = ? AND refresh_token=?`,
		newAccess, newExp, newRefreshExp, req.UserID, req.RefreshToken)
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