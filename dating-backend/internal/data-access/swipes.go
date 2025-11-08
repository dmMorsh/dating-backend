package data_access

import (
	"dating-backend/internal/models"
	"dating-backend/internal/utils"
	"math"
)

// UpsertSwipe puts or updates a swipe record
func UpsertSwipe(userID, targetID int64, action string) error {
	_, err := DB.Exec(`
		INSERT OR REPLACE INTO swipes (user_id, target_id, action)
		VALUES (?, ?, ?)
	`, userID, targetID, action)
	return err
}

// HasLiked checks if userID has liked targetID
func HasLiked(userID, targetID int64) (bool, error) {
	var cnt int
	err := DB.QueryRow(`
		SELECT COUNT(*) FROM swipes
		WHERE user_id = ? AND target_id = ? AND action = 'like'
	`, userID, targetID).Scan(&cnt)
	if err != nil {
		return false, err
	}
	return cnt > 0, nil
}

// UpsertSwipe and HasLiked are thin wrappers for database operations related
// to swipe state. Keeping them in data-access centralizes DB code and makes
// higher-level handlers easier to test and reason about.
func GetUserFollowers(userID int64) ([]models.User, error) {
	rows, err := DB.Query(`
		SELECT 
		l1.user_id, u.name, u.birthday, u.photo_url, u.bio
		FROM swipes l1
		JOIN users u ON l1.user_id = u.id
		WHERE 
			l1.target_id = ? 
			AND l1.action = 'like'
			AND l1.user_id NOT IN (
				SELECT l2.target_id
				FROM swipes l2
				WHERE l2.user_id = ?
				)
		`, userID, userID)	
		
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var followers []models.User
	for rows.Next() {
		var follower models.User
		if err := rows.Scan(&follower.ID, &follower.Name, &follower.Birthday, &follower.PhotoURL, &follower.Bio,); err != nil {

			return nil, err
		}
		follower.Age = utils.GetAge(&follower.Birthday.Time)
		followers = append(followers, follower)
	}
	return followers, nil
}

// GetSwipeCandidates returns a list of users that the given user has not swiped on yet,
// applying optional filters from SimpleFilter.
func GetSwipeCandidates(userID int64, f *models.SimpleFilter) ([]models.User, error) {
	query := `
	SELECT
		u.id, u.username, u.name, u.gender, u.birthday,
		u.interested_in, u.bio, u.photo_url, u.location,
		u.latitude, u.longitude, u.created_at, u.last_active
	FROM users u
	JOIN user_locations ul ON ul.id = u.id
	LEFT JOIN swipes s ON s.target_id = u.id AND s.user_id = ?
	WHERE u.id != ?
	  AND s.id IS NULL
	`
	args := []any{userID, userID}

	// --- dinamic filters ---
	var lat1, lon1 float64
	var useGeo bool
	if f.Latitude != nil && f.Longitude != nil && f.MaxDistanceKm != nil {
		useGeo = true
		const R = 6371.0
		lat1 = *f.Latitude
		lon1 = *f.Longitude
		dist := *f.MaxDistanceKm

		// in degrees
		dLat := dist / R * (180 / math.Pi)
		dLon := dist / (R * math.Cos(lat1*math.Pi/180)) * (180 / math.Pi)

		minLat := lat1 - dLat
		maxLat := lat1 + dLat
		minLon := lon1 - dLon
		maxLon := lon1 + dLon

		query += " AND ul.min_lat >= ? AND ul.max_lat <= ? AND ul.min_lon >= ? AND ul.max_lon <= ?"
		args = append(args, minLat, maxLat, minLon, maxLon)
	}

	if f.Gender != nil && *f.Gender != "" {
		query += " AND u.gender = ?"
		args = append(args, *f.Gender)
	}

	if f.MinAge != nil && f.MaxAge != nil {
		query += " AND u.birthday IS NOT NULL AND ((julianday('now') - julianday(u.birthday)) / 365.25) BETWEEN ? AND ?"
		args = append(args, *f.MinAge, *f.MaxAge)
	}

	if f.HasPhoto != nil && *f.HasPhoto {
		query += " AND u.photo_url != ''"
	}

	if f.InterestedIn != nil && *f.InterestedIn != "" {
		query += " AND u.interested_in LIKE ?"
		args = append(args, "%"+*f.InterestedIn+"%")
	}

	if f.LastSeenID != nil {
		query += " AND u.id > ?"
		args = append(args, *f.LastSeenID)
	}

	// --- sort and limits ---
	query += `
	ORDER BY u.id ASC
	LIMIT ?
	`
	args = append(args, f.PageSize)

	rows, err := DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var candidates []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(
			&u.ID, &u.Username, &u.Name, &u.Gender, &u.Birthday,
			&u.InterestedIn, &u.Bio, &u.PhotoURL, &u.Location,
			&u.Latitude, &u.Longitude, &u.CreatedAt, &u.LastActive,
		); err != nil {
			return nil, err
		}

		u.Age = utils.GetAge(&u.Birthday.Time)

		if useGeo && u.Latitude != nil && u.Longitude != nil {
			dist := haversine(lat1, lon1, *u.Latitude, *u.Longitude)
			if dist > *f.MaxDistanceKm {
				continue // skip users outside the radius
			}
		}
		u.Latitude = nil 
		u.Longitude= nil

		candidates = append(candidates, u)
	}

	return candidates, nil
}

func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371.0 // Earth radius in km
	dLat := (lat2 - lat1) * math.Pi / 180.0
	dLon := (lon2 - lon1) * math.Pi / 180.0
	lat1R := lat1 * math.Pi / 180.0
	lat2R := lat2 * math.Pi / 180.0

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(lat1R)*math.Cos(lat2R)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

// Only for testing purposes
func ClearSwipesForUser(userID int64) (error) {
	_, err := DB.Exec(`DELETE FROM swipes WHERE user_id = ?`,
	userID)
	if err != nil {
		return err
	}
	return nil
}