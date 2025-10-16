package data_access

import (
	"database/sql"
	"fmt"
)

func migrate(db *sql.DB) error {
	// Проверяем, есть ли колонка "birthday"
	rows, err := db.Query(`PRAGMA table_info(users);`)
	if err != nil {
		return err
	}
	defer rows.Close()

	hasBirthday := false
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt sql.NullString
		_ = rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk)
		if name == "birthday" {
			hasBirthday = true
			break
		}
	}

	if !hasBirthday {
		fmt.Println("Migrating: adding 'birthday' column...")
		_, err = db.Exec(`ALTER TABLE users ADD COLUMN birthday DATETIME;`)
		if err != nil {
			return fmt.Errorf("failed to add birthday column: %w", err)
		}
	}

	return nil
}