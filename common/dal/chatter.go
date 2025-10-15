package dal

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

func InsertChatter(db *sql.DB, username, name string) (*Chatter, error) {

	if username == "" {
		return nil, fmt.Errorf("username cannot be empty")
	}

	stmt := `INSERT INTO chatters (username, name) VALUES (?, ?)`
	result, err := db.Exec(stmt, username, name)
	if err != nil {
		return nil, err
	}

	chatterID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	chatter := &Chatter{
		ID:       chatterID,
		Username: username,
		Name:     name,
	}

	return chatter, nil
}

func GetChatterByUsername(db *sql.DB, username string) (*Chatter, error) {

	if username == "" {
		return nil, fmt.Errorf("username cannot be empty")
	}

	stmt := `SELECT id, username, name FROM chatters WHERE username = ?`
	var chatter Chatter
	err := db.QueryRow(stmt, username).Scan(&chatter.ID, &chatter.Username, &chatter.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("chatter with username '%s' not found", username)
		}
		return nil, err
	}

	return &chatter, nil
}

func TotalChatters(db *sql.DB) (int64, error) {
	query := `SELECT COUNT(*) FROM chatters`

	var count int64
	err := db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}