package data_access

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func setupInMemoryDB(t *testing.T) func() {
    var err error
    DB, err = sql.Open("sqlite", ":memory:")
    if err != nil {
        t.Fatalf("open db: %v", err)
    }

    // create tables like in InitDB
    stmts := []string{
        `CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT, username TEXT UNIQUE NOT NULL, password TEXT NOT NULL);`,
        `CREATE TABLE swipes (id INTEGER PRIMARY KEY AUTOINCREMENT, user_id INTEGER NOT NULL, target_id INTEGER NOT NULL, action TEXT CHECK(action IN ('like', 'dislike')) NOT NULL, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, UNIQUE(user_id, target_id));`,
        `CREATE TABLE chats (id INTEGER PRIMARY KEY AUTOINCREMENT, user1_id INTEGER NOT NULL, user2_id INTEGER NOT NULL, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, UNIQUE(user1_id, user2_id));`,
    }
    for _, s := range stmts {
        if _, err := DB.Exec(s); err != nil {
            t.Fatalf("create table: %v", err)
        }
    }

    return func() {
        DB.Close()
    }
}

func TestUpsertAndHasLiked(t *testing.T) {
    tear := setupInMemoryDB(t)
    defer tear()

    // insert two users
    res, err := DB.Exec(`INSERT INTO users (username, password) VALUES (?, ?)`, "a", "p")
    if err != nil { t.Fatalf("insert user: %v", err) }
    _, _ = res.LastInsertId()
    res, err = DB.Exec(`INSERT INTO users (username, password) VALUES (?, ?)`, "b", "p")
    if err != nil { t.Fatalf("insert user: %v", err) }

    // assume ids 1 and 2
    if err := UpsertSwipe(1, 2, "like"); err != nil { t.Fatalf("upsert: %v", err) }
    liked, err := HasLiked(1, 2)
    if err != nil { t.Fatalf("hasliked: %v", err) }
    if !liked { t.Fatalf("expected liked true") }

    // update action
    if err := UpsertSwipe(1, 2, "dislike"); err != nil { t.Fatalf("upsert: %v", err) }
    liked, err = HasLiked(1, 2)
    if err != nil { t.Fatalf("hasliked: %v", err) }
    if liked { t.Fatalf("expected liked false after dislike") }
}

func TestCreateOrGetChat_Uniqueness(t *testing.T) {
    tear := setupInMemoryDB(t)
    defer tear()

    // create users
    DB.Exec(`INSERT INTO users (username, password) VALUES (?, ?)`, "a", "p")
    DB.Exec(`INSERT INTO users (username, password) VALUES (?, ?)`, "b", "p")

    id1, err := CreateOrGetChat(1, 2)
    if err != nil { t.Fatalf("create: %v", err) }
    id2, err := CreateOrGetChat(2, 1)
    if err != nil { t.Fatalf("get: %v", err) }
    if id1 != id2 {
        t.Fatalf("expected same chat id, got %d and %d", id1, id2)
    }
}
