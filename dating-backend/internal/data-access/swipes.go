package data_access

import (
	"dating-backend/internal/models"
	// "time"
)

// Добавить лайк
// func AddLike(fromUserID, toUserID int64) error {
// 	_, err := DB.Exec(`
// 		INSERT OR IGNORE INTO likes (from_user, to_user, created_at)
// 		VALUES (?, ?, ?)
// 	`, fromUserID, toUserID, time.Now())
// 	return err
// }

// Проверить, есть ли взаимный лайк
// func IsMatch(userA, userB int64) (bool, error) {
// 	var exists bool
// 	err := DB.QueryRow(`
// 		SELECT EXISTS(
// 			SELECT 1 FROM likes AS l1
// 			JOIN likes AS l2 ON l1.from_user = l2.to_user AND l1.to_user = l2.from_user
// 			WHERE l1.from_user = ? AND l1.to_user = ?
// 		)
// 	`, userA, userB).Scan(&exists)
// 	return exists, err
// }

// Получить всех мэтчей для пользователя
func GetMatches(userID int64) ([]models.User, error) {
	// rows, err := DB.Query(`
	// 	SELECT u.id, u.username, u.name, u.bio, u.photo_url
	// 	FROM users u
	// 	WHERE u.id IN (
	// 		SELECT l2.from_user
	// 		FROM likes l1
	// 		JOIN likes l2 ON l1.from_user = l2.to_user AND l1.to_user = l2.from_user
	// 		WHERE l1.from_user = ?
	// 	)
	// `, userID)
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