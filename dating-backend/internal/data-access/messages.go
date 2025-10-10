package data_access

import (
	"dating-backend/internal/models"
)

// 1. Сохранить сообщение
func SaveMessage(msg *models.Message) error {
	_, err := DB.Exec(`
		INSERT INTO messages (chat_id, sender_id, receiver_id, content, is_read, created_at)
		VALUES (?, ?, ?, ?, 0, datetime('now'))
	`, msg.ChatID, msg.SenderID, msg.ReceiverID, msg.Content)
	return err
}

// 2. Получить непрочитанные
func GetUnreadMessages(userID int64) ([]models.Message, error) {
	rows, err := DB.Query(`
		SELECT id, sender_id, receiver_id, content, is_read, created_at
		FROM messages
		WHERE receiver_id = ? AND is_read = 0
		ORDER BY created_at ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []models.Message
	for rows.Next() {
		var m models.Message
		if err := rows.Scan(&m.ID, &m.SenderID, &m.ReceiverID, &m.Content, &m.IsRead, &m.CreatedAt); err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}
	return msgs, nil
}

// 3. Пометить как прочитанные
func MarkMessagesAsRead(userID int64) error {
	_, err := DB.Exec(`UPDATE messages SET is_read = 1 WHERE receiver_id = ? AND is_read = 0`, userID)
	return err
}

func GetMessagesBetweenUsers(user1ID, user2ID int64) ([]models.Message, error) {
	rows, err := DB.Query(`
		SELECT id, sender_id, receiver_id, content, created_at
		FROM messages
		WHERE (sender_id = ? AND receiver_id = ?)
		   OR (sender_id = ? AND receiver_id = ?)
		ORDER BY created_at ASC`,
		user1ID, user2ID, user2ID, user1ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []models.Message
	for rows.Next() {
		var m models.Message
		if err := rows.Scan(&m.ID, &m.SenderID, &m.ReceiverID, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}
	return msgs, nil
}

func CreateOrGetChat(userA, userB int64) (int64, error) {
	// Normalize order so (a,b) and (b,a) map to same chat
	if userA > userB {
		userA, userB = userB, userA
	}

	// Use a transaction and INSERT OR IGNORE to avoid race conditions
	tx, err := DB.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// Try to insert (will fail silently on unique constraint)
	_, err = tx.Exec(`INSERT OR IGNORE INTO chats (user1_id, user2_id) VALUES (?, ?)`, userA, userB)
	if err != nil {
		return 0, err
	}

	var chatID int64
	err = tx.QueryRow(`
		SELECT id FROM chats
		WHERE (user1_id = ? AND user2_id = ?)
		   OR (user1_id = ? AND user2_id = ?)
	`, userA, userB, userB, userA).Scan(&chatID)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return chatID, nil
}
