package data_access

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

var DB *sql.DB


func InitDB() {
	var err error
	DB, err = sql.Open("sqlite", "./dating.db")
	if err != nil {
		log.Fatal(err)
	}

	createUsers := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		name TEXT,
		gender TEXT,
		age INTEGER,
		interested_in TEXT,
		bio TEXT,
		photo_url TEXT,
		location TEXT,
		latitude REAL,
		longitude REAL,
		created_at TEXT DEFAULT CURRENT_TIMESTAMP,
		last_active TEXT
	);`
	
	createLikes := `
	CREATE TABLE IF NOT EXISTS likes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		from_user INTEGER NOT NULL,
		to_user INTEGER NOT NULL,
		is_match BOOLEAN
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
		FOREIGN KEY(user_id) REFERENCES users(id)
	);`

	createMessages := `
	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		match_id INTEGER,
		sender_id INTEGER,
		text TEXT,
		sent_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = DB.Exec(createUsers)
	if err != nil {
		log.Fatal(err)
	}
	_, err = DB.Exec(createLikes)
	if err != nil {
		log.Fatal(err)
	}
	_, err = DB.Exec(createSessions)
	if err != nil {
		log.Fatal(err)
	}
	_, err = DB.Exec(createMessages)
	if err != nil {
		log.Fatal(err)
	}
}