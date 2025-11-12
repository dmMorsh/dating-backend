package models

import (
	"time"
)

type Chat struct {
	ID              int64     `json:"id"`
	User1ID         int64     `json:"user1_id"`
	User2ID         int64     `json:"user2_id"`
	CreatedAt       time.Time `json:"created_at"`
	LastMessage     string    `json:"last_message,omitempty"`      // Optional field for last message preview
	LastMessageTime time.Time `json:"last_message_time,omitempty"` // Optional field for last message time
	LastMessageUser int64	  `json:"last_message_user,omitempty"` // Optional field for last message user ID
	IsRead			bool      `json:"is_read"`
}