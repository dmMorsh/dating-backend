package data_access

import (
	"dating-backend/internal/models"
	"time"
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

func GetChatsForUser(userID int64) ([]models.Chat, error) {
	rows, err := DB.Query(`
		SELECT 
		c.id, 
		c.user1_id, 
		c.user2_id, 
		c.created_at, 
		IFNULL(m.last_message, '') AS last_message, 
		IFNULL(m.last_message_time, '') AS last_message_time,
		IFNULL(m.last_message_user, 0) AS last_message_user
		FROM chats c
		LEFT JOIN (
			SELECT 
			chat_id, 
			content AS last_message, 
			sender_id AS last_message_user,
			MAX(created_at) AS last_message_time
			FROM messages
			GROUP BY chat_id
		) m ON c.id = m.chat_id
		WHERE c.user1_id = ? OR c.user2_id = ?
	`, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chats []models.Chat
	for rows.Next() {
		var c models.Chat
		var lastMessageTimeStr string
		if err := rows.Scan(&c.ID, &c.User1ID, &c.User2ID, &c.CreatedAt, &c.LastMessage,
			&lastMessageTimeStr, &c.LastMessageUser); err != nil {
			return nil, err
		}
		if lastMessageTimeStr != "" {
			parsedTime, err := time.Parse("2006-01-02 15:04:05", lastMessageTimeStr)
			if err != nil {
				return nil, err
			}
			c.LastMessageTime = parsedTime
		}
		chats = append(chats, c)
	}
	return chats, nil
}

func GetMessagesForChat(chatID int64) ([]models.Message, error) {
	rows, err := DB.Query(`
		SELECT id, sender_id, receiver_id, content, is_read, created_at
		FROM messages
		WHERE chat_id = ?
		ORDER BY created_at ASC
	`, chatID)
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