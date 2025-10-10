package data_access

import (
	"dating-backend/models"
	"time"
)

// Добавить лайк
func AddLike(fromUserID, toUserID int64) error {
	_, err := DB.Exec(`
		INSERT OR IGNORE INTO likes (from_user_id, to_user_id, created_at)
		VALUES (?, ?, ?)
	`, fromUserID, toUserID, time.Now())
	return err
}

// Проверить, есть ли взаимный лайк
func IsMatch(userA, userB int64) (bool, error) {
	var exists bool
	err := DB.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM likes AS l1
			JOIN likes AS l2 ON l1.from_user_id = l2.to_user_id AND l1.to_user_id = l2.from_user_id
			WHERE l1.from_user_id = ? AND l1.to_user_id = ?
		)
	`, userA, userB).Scan(&exists)
	return exists, err
}

// Получить всех мэтчей для пользователя
func GetMatches(userID int64) ([]models.User, error) {
	rows, err := DB.Query(`
		SELECT u.id, u.username, u.name, u.bio, u.photo_url
		FROM users u
		WHERE u.id IN (
			SELECT l2.from_user_id
			FROM likes l1
			JOIN likes l2 ON l1.from_user_id = l2.to_user_id AND l1.to_user_id = l2.from_user_id
			WHERE l1.from_user_id = ?
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