package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	data_access "dating-backend/internal/data-access"
	middleware "dating-backend/internal/middleware"
	"dating-backend/internal/utils"
)

// GET /me
// Retrieves the profile of the authenticated user.
// Example response:
// {
//	 "id": 1,
//	 "username": "johndoe",
//	 "name": "John Doe",
//	 "bio": "Hello!",
//	 "photo_url": "http://example.com/photo.jpg",
//	 ...
// }
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

	u.Password = "" // Hide password
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(u)
}

// GET /user/{id}
// Retrieves the profile of a user by their ID.
// Example response:
// {
//	 "id": 2,
//	 "username": "janedoe",
//	 "name": "Jane Doe",
//	 "bio": "Hi there!",
//	 "photo_url": "http://example.com/photo2.jpg",
//	 ...
// }
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

	u.Age = utils.GetAge(u.Birthday)
	u.Birthday = nil // Hide birthday
	u.Password = ""
	u.Longitude = nil // Hide precise location
	u.Latitude = nil
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(u)
}


type UpdateProfileRequest struct {
	Name         	*string  `json:"name,omitempty"`
	Gender       	*string  `json:"gender,omitempty"`
	Birthday		*utils.JSONDate	`json:"birthday,omitempty"`
	InterestedIn 	*string  `json:"interested_in,omitempty"`
	Bio          	*string  `json:"bio,omitempty"`
	PhotoURL     	*string  `json:"photo_url,omitempty"`
	Location     	*string  `json:"location,omitempty"`
	Latitude     	*float64 `json:"latitude,omitempty"`
	Longitude    	*float64 `json:"longitude,omitempty"`
}

// PUT /me
// Updates the profile of the authenticated user.
// Expects a JSON body with fields to update.
// Example request body:
// {
//	 "name": "New Name",
//	 "bio": "Updated bio",
//	 "photo_url": "http://example.com/newphoto.jpg",
//	 ...
// }
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

	// Apply changes only if they have arrived in the request
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
		u.Location = req.Location
	}
	if req.Latitude != nil && *req.Latitude != 0.0 {
		u.Latitude = req.Latitude
	}
	if req.Longitude != nil && *req.Longitude != 0.0 {
		u.Longitude = req.Longitude
	}

	if err := data_access.UpdateUser(u); err != nil {
		http.Error(w, "failed to update", http.StatusInternalServerError)
		return
	}

	u.Password = ""
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(u)
}