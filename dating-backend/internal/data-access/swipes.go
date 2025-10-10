package data_access

import (
	"dating-backend/internal/models"
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