package data_access

import (
	"dating-backend/internal/models"
	"dating-backend/internal/utils"
)

// Получить всех мэтчей для пользователя
func GetMatches(userID int64) ([]models.User, error) {
	rows, err := DB.Query(`
		SELECT u.id, u.username, u.name, u.bio, u.photo_url
		FROM users u
		WHERE u.id IN (
			SELECT l2.user_id
			FROM swipes l1
			JOIN swipes l2 ON l1.user_id = l2.target_id AND l1.target_id = l2.user_id
			AND l1.action = 'like' AND l2.action = 'like'
			WHERE l1.user_id = ?
		)
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var matches []models.User
	for rows.Next() {
		var u models.User
		rows.Scan(&u.ID, &u.Username, &u.Name, &u.Bio, &u.PhotoURL)
		matches = append(matches, u)
	}
	return matches, nil
}

// UpsertSwipe вставляет или обновляет запись о свайпе
func UpsertSwipe(userID, targetID int64, action string) error {
	_, err := DB.Exec(`
		INSERT OR REPLACE INTO swipes (user_id, target_id, action)
		VALUES (?, ?, ?)
	`, userID, targetID, action)
	return err
}

// HasLiked проверяет, поставил ли userID лайк targetID
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
		follower.Age = utils.GetAge(follower.Birthday)
		followers = append(followers, follower)
	}
	return followers, nil
}

func GetSwipeCandidates(userID int64, f *models.SimpleFilter) ([]models.User, error) {
	query := `
	SELECT
		u.id, u.username, u.name, u.gender, u.birthday,
		u.interested_in, u.bio, u.photo_url, u.location,
		u.latitude, u.longitude, u.created_at, u.last_active
	FROM users u
	LEFT JOIN swipes s ON s.target_id = u.id AND s.user_id = ?
	WHERE u.id != ?
	  AND s.id IS NULL
	`
	args := []any{userID, userID}

	// --- динамические фильтры ---
	if f.Gender != nil && *f.Gender != "" {
		query += " AND u.gender = ?"
		args = append(args, *f.Gender)
	}

	if f.MinAge != nil && f.MaxAge != nil {
		query += " AND (strftime('%Y', 'now') - strftime('%Y', u.birthday)) BETWEEN ? AND ?"
		args = append(args, *f.MinAge, *f.MaxAge)
	}

	if f.HasPhoto != nil && *f.HasPhoto {
		query += " AND u.photo_url != ''"
	}

	if f.InterestedIn != "" {
		query += " AND u.interested_in LIKE ?"
		args = append(args, "%"+f.InterestedIn+"%")
	}

	if f.LastSeenID != nil {
		query += " AND u.id > ?"
		args = append(args, *f.LastSeenID)
	}

	// --- сортировка и лимиты ---
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

	
//TODO hide lat/lon
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
		u.Age = utils.GetAge(u.Birthday)
		candidates = append(candidates, u)
	}

	return candidates, nil
}

func ClearSwipesForUser(userID int64) (error) {
	_, err := DB.Exec(`DELETE FROM swipes WHERE user_id = ?`,
	userID)
	if err != nil {
		return err
	}
	return nil
}