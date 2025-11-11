package data_access

import (
	"database/sql"
	"dating-backend/internal/logging"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB() {
	var err error
	DB, err = sql.Open("sqlite", "./dating.db")
	if err != nil {
		logging.Log.Fatalw("failed to open DB", "err", err)
	}

	createUsers := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		name TEXT,
		gender TEXT,
		interested_in TEXT,
		bio TEXT,
		photo_url TEXT,
		location TEXT,
		latitude REAL,
		longitude REAL,
		created_at TEXT DEFAULT CURRENT_TIMESTAMP,
		last_active TEXT
	);`
	
	createSwipes := `
	CREATE TABLE IF NOT EXISTS swipes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		target_id INTEGER NOT NULL,
		action TEXT CHECK(action IN ('like', 'dislike')) NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(user_id, target_id)
	);`

	createSessions := `
	CREATE TABLE IF NOT EXISTS sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		device_id TEXT NOT NULL,
		access_token TEXT NOT NULL UNIQUE,
		refresh_token TEXT NOT NULL UNIQUE,
		access_expires DATETIME NOT NULL,
		refresh_expires DATETIME NOT NULL,
		UNIQUE(user_id, device_id)
		FOREIGN KEY(user_id) REFERENCES users(id)
	);`

	createChats := `
	CREATE TABLE IF NOT EXISTS chats (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user1_id INTEGER NOT NULL,
		user2_id INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(user1_id, user2_id)
	);`

	createMessages := `
	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		chat_id INTEGER NOT NULL,
		sender_id INTEGER NOT NULL,
		receiver_id INTEGER NOT NULL,
		content TEXT NOT NULL,
		is_read BOOLEAN DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (sender_id) REFERENCES users(id),
		FOREIGN KEY (receiver_id) REFERENCES users(id)
	);`

	createUserLocations := `
	CREATE VIRTUAL TABLE IF NOT EXISTS user_locations USING rtree(
		id,
		min_lat,
		max_lat,
		min_lon,
		max_lon
	);`

	_, err = DB.Exec(createUsers)
	if err != nil {
		logging.Log.Fatalw("failed to exec createUsers", "err", err)
	}
	_, err = DB.Exec(createSwipes)
	if err != nil {
		logging.Log.Fatalw("failed to exec createSwipes", "err", err)
	}
	_, err = DB.Exec(createSessions)
	if err != nil {
		logging.Log.Fatalw("failed to exec createSessions", "err", err)
	}
	_, err = DB.Exec(createChats)
	if err != nil {
		logging.Log.Fatalw("failed to exec createChats", "err", err)
	}
	_, err = DB.Exec(createMessages)
	if err != nil {
		logging.Log.Fatalw("failed to exec createMessages", "err", err)
	}
	_, err = DB.Exec(createUserLocations)
	if err != nil {
		logging.Log.Fatalw("failed to exec createUserLocations", "err", err)
	}

	// Run migrations
	if err := migrate(DB); err != nil {
		logging.Log.Fatalw("Migration failed", "err", err)
    }
}