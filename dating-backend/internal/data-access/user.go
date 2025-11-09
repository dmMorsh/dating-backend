package data_access

import (
	"database/sql"
	"dating-backend/internal/models"
	"fmt"
	"log"
)

// GetUserByID retrieves a user by their ID.
func GetUserByID(id int64) (*models.User, error) {
	u := &models.User{}
	row := DB.QueryRow(`
		SELECT
		id,
		username,
		IFNULL(name, ''),
		IFNULL(gender, ''),
		birthday,
		IFNULL(interested_in, ''),
		IFNULL(bio, ''),
		IFNULL(photo_url, ''),
		IFNULL(location, ''),
		IFNULL(latitude, 0),
		IFNULL(longitude, 0),
		IFNULL(created_at, ''),
		IFNULL(last_active, '')
	FROM users WHERE id = ?`, id)

	var b models.SQLiteDate
	err := row.Scan(&u.ID, &u.Username, &u.Name, &u.Gender, &b,
		&u.InterestedIn, &u.Bio, &u.PhotoURL, &u.Location, &u.Latitude, 
		&u.Longitude, &u.CreatedAt, &u.LastActive)
	if err == sql.ErrNoRows {
		log.Printf("data-access: GetUserByID not found id=%d", id)
		return nil, fmt.Errorf("not found")
	}
	if err != nil {
		log.Printf("data-access: GetUserByID scan error id=%d: %v", id, err)
		return nil, err
	}

	if !b.Time.IsZero() {
		u.Birthday = &b
	}

	return u, nil
}

// UpdateUser updates the user's profile information.
func UpdateUser(u *models.User) error {
	_, err := DB.Exec(`
		UPDATE users SET 
		name=?,
		gender=?,
		birthday=?,
		interested_in=?,
		bio=?,
		photo_url=?,
		location=?,
		latitude=?,
		longitude=?,
		last_active=CURRENT_TIMESTAMP
		WHERE id=?`,
		u.Name, u.Gender, u.Birthday, u.InterestedIn, u.Bio, u.PhotoURL, 
		u.Location, u.Latitude, u.Longitude, u.ID,
	)
	if err != nil {
		log.Printf("data-access: UpdateUser error id=%d: %v", u.ID, err)
	}
	return err
}

func UpdateUserLocationIndex(userID int64, lat, lon float64) error {
	_, err := DB.Exec(`
	    INSERT OR REPLACE INTO user_locations 
	    (id, min_lat, max_lat, min_lon, max_lon)
	    VALUES (?, ?, ?, ?, ?)`,
		userID, lat, lat, lon, lon)
	if err != nil {
		log.Printf("data-access: UpdateUserLocationIndex error id=%d: %v", userID, err)
	}
	return err
}