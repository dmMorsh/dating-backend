package data_access

import (
	"database/sql"
	"dating-backend/models"
	"fmt"
)

func GetUserByID(id int64) (*models.User, error) {
	u := &models.User{}
	row := DB.QueryRow(`
		SELECT
		id,
		username,
		IFNULL(name, ''),
		IFNULL(gender, ''),
		IFNULL(age, 0),
		IFNULL(interested_in, ''),
		IFNULL(bio, ''),
		IFNULL(photo_url, ''),
		IFNULL(location, ''),
		IFNULL(latitude, 0),
		IFNULL(longitude, 0),
		IFNULL(created_at, ''),
		IFNULL(last_active, '')
	FROM users WHERE id = ?`, id)

	err := row.Scan(&u.ID, &u.Username, &u.Name, &u.Gender, &u.Age, 
		&u.InterestedIn, &u.Bio, &u.PhotoURL, &u.Location, &u.Latitude, 
		&u.Longitude, &u.CreatedAt, &u.LastActive)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("not found")
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

func UpdateUser(u *models.User) error {
	_, err := DB.Exec(`
		UPDATE users SET 
		name=?, 
		gender=?, 
		age=?, 
		interested_in=?, 
		bio=?, 
		photo_url=?, 
		location=?, 
		latitude=?, 
		longitude=?, 
		last_active=CURRENT_TIMESTAMP
		WHERE id=?`,
		u.Name, u.Gender, u.Age, u.InterestedIn, u.Bio, u.PhotoURL, 
		u.Location, u.Latitude, u.Longitude, u.ID,
	)
	return err
}