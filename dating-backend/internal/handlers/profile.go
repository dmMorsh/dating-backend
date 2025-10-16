package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	data_access "dating-backend/internal/data-access"
	middleware "dating-backend/internal/middleware"
	models "dating-backend/internal/models"
)

// GET /me
func GetMyProfileHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	u, err := data_access.GetUserByID(userID)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	u.Password = "" // на всякий случай
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(u)
}

// GET /user/{id}
func GetUserHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/user/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	u, err := data_access.GetUserByID(id)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	u.Password = ""
	u.Longitude = 0 // прячем координаты
	u.Latitude = 0
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(u)
}

// PUT /me
type UpdateProfileRequest struct {
	Name         *string  `json:"name,omitempty"`
	Gender       *string  `json:"gender,omitempty"`
	Birthday   *JSONDate  `json:"birthday,omitempty"`
	InterestedIn *string  `json:"interested_in,omitempty"`
	Bio          *string  `json:"bio,omitempty"`
	PhotoURL     *string  `json:"photo_url,omitempty"`
	Location     *string  `json:"location,omitempty"`
	Latitude     *float64 `json:"latitude,omitempty"`
	Longitude    *float64 `json:"longitude,omitempty"`
}

func UpdateProfileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	u, err := data_access.GetUserByID(userID)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	// Применяем изменения только если пришли (NULLable поля)
	if req.Name != nil {
		u.Name = strings.TrimSpace(*req.Name)
	}
	if req.Gender != nil {
		u.Gender = *req.Gender
	}
	if req.Birthday != nil {
		t := req.Birthday.Time()
		u.Birthday = &t
	}
	if req.InterestedIn != nil {
		u.InterestedIn = *req.InterestedIn
	}
	if req.Bio != nil {
		u.Bio = *req.Bio
	}
	if req.PhotoURL != nil {
		u.PhotoURL = *req.PhotoURL
	}
	if req.Location != nil {
		u.Location = *req.Location
	}
	if req.Latitude != nil && *req.Latitude != 0.0 {
		u.Latitude = *req.Latitude
	}
	if req.Longitude != nil && *req.Longitude != 0.0 {
		u.Longitude = *req.Longitude
	}

	if err := data_access.UpdateUser(u); err != nil {
		http.Error(w, "failed to update", http.StatusInternalServerError)
		return
	}

	u.Password = ""
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(u)
}

func ProfilesHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := data_access.DB.Query("SELECT id, username, bio, photo_url FROM users")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var profiles []models.User
	for rows.Next() {
		var u models.User
		err := rows.Scan(&u.ID, &u.Username, &u.Bio, &u.PhotoURL)
		if err != nil {
			continue
		}
		profiles = append(profiles, u)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profiles)
}



type JSONDate time.Time

func (jt *JSONDate) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "" {
		return nil
	}

	// пробуем несколько форматов
	layouts := []string{
		time.RFC3339,              // 2006-01-02T15:04:05Z07:00
		"2006-01-02T15:04:05",     // без зоны
		"2006-01-02",              // только дата
	}

	var t time.Time
	var err error
	for _, layout := range layouts {
		t, err = time.Parse(layout, s)
		if err == nil {
			*jt = JSONDate(t)
			return nil
		}
	}
	return err
}

func (jt JSONDate) Time() time.Time {
	return time.Time(jt)
}