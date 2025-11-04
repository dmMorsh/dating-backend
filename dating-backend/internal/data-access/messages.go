package data_access

import (
	"dating-backend/internal/models"
	"time"
)

func SaveMessage(msg *models.Message) (int64, error) {
	res, _ := DB.Exec(`
		INSERT INTO messages (chat_id, sender_id, receiver_id, content, is_read, created_at)
		VALUES (?, ?, ?, ?, 0, datetime('now'))
	`, msg.ChatID, msg.SenderID, msg.ReceiverID, msg.Content)
	return res.LastInsertId()
}

func CreateOrGetChat(userA, userB int64) (bool, int64, error) {
	// Normalize order so (a,b) and (b,a) map to same chat
	if userA > userB {
		userA, userB = userB, userA
	}

	// Use a transaction and INSERT OR IGNORE to avoid race conditions
	tx, err := DB.Begin()
	if err != nil {
		return false, 0, err
	}
	defer tx.Rollback()

	// Try to insert (will fail silently on unique constraint)
	res, err := tx.Exec(`INSERT OR IGNORE INTO chats (user1_id, user2_id) VALUES (?, ?)`, userA, userB)
	if err != nil {
		return false, 0, err
	}
	rowsAffected, _ := res.RowsAffected()
	createdNew := rowsAffected > 0

	var chatID int64
	err = tx.QueryRow(`
		SELECT id FROM chats
		WHERE (user1_id = ? AND user2_id = ?)
		   OR (user1_id = ? AND user2_id = ?)
	`, userA, userB, userB, userA).Scan(&chatID)
	if err != nil {
		return false, 0, err
	}

	if err := tx.Commit(); err != nil {
		return false, 0, err
	}
	return createdNew, chatID, nil
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
		IFNULL(m.last_message_user, 0) AS last_message_user,
		m.is_read
		FROM chats c
		LEFT JOIN (
			SELECT 
			chat_id, 
			content AS last_message, 
			sender_id AS last_message_user,
			MAX(created_at) AS last_message_time,
			is_read
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
			&lastMessageTimeStr, &c.LastMessageUser, &c.IsRead); err != nil {
			return nil, err
		}
		if c.LastMessageUser == userID || c.LastMessageUser == 0 {
			c.IsRead = true
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

func GetMessagesForChat(chatID int64, beforeID, afterID *int64, limit int) ([]models.Message, error) {
	// ✅ Case 1: First load — get LAST MESSAGES
	if beforeID == nil && afterID == nil {
		query := `
			SELECT * FROM (
				SELECT id, sender_id, receiver_id, content, is_read, created_at
				FROM messages
				WHERE chat_id = ?
				ORDER BY id DESC
				LIMIT ?
			) sub
			ORDER BY id ASC
		`

		rows, err := DB.Query(query, chatID, limit)
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

	// ✅ Case 2/3: old or new messages
	query := `
		SELECT id, sender_id, receiver_id, content, is_read, created_at
		FROM messages
		WHERE chat_id = ?
	`
	args := []interface{}{chatID}

	if beforeID != nil {
		query += " AND id < ?"
		args = append(args, *beforeID)
	}

	if afterID != nil {
		query += " AND id > ?"
		args = append(args, *afterID)
	}

	if beforeID != nil {
		// load older messages
		query = `
			SELECT * FROM (
				` + query + `
				ORDER BY id DESC
				LIMIT ?
			) sub ORDER BY id ASC
		`
	} else {
		// load newer messages
		query += " ORDER BY id ASC LIMIT ?"
	}

	args = append(args, limit)

	rows, err := DB.Query(query, args...)
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

func MarkMessagesAsReadForChat(chatID int64, userID int64) (bool, error) {
	_, err := DB.Exec(`
		UPDATE messages SET is_read = 1
		WHERE chat_id = ? AND receiver_id = ? AND is_read = 0
	`, chatID, userID)
	if err != nil {
		return false, err
	}
	return true, nil
}

func MarkMessagesAsRead(MessageIDs []int64) (bool, error) {
	query := "UPDATE messages SET is_read = TRUE WHERE id IN ("
	args := make([]interface{}, len(MessageIDs))
	for i, id := range MessageIDs {
		if i > 0 {
			query += ","
		}
		query += "?"
		args[i] = id
	}
	query += ")"

	_, err := DB.Exec(query, args...)
	if err != nil {
		return false, err
	}
	return true, nil
}