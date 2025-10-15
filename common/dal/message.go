package dal

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

func InsertMessage(db *sql.DB, userID, roomID int64, content string) (*Message, error) {
	stmt := `INSERT INTO messages (userId, roomId, content) VALUES (?, ?, ?)`
	result, err := db.Exec(stmt, userID, roomID, content)
	if err != nil {
		return nil, err
	}

	messageID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	selectStmt := `SELECT id, userId, roomId, content, timestamp FROM messages WHERE id = ?`
	var msg Message
	err = db.QueryRow(selectStmt, messageID).Scan(&msg.ID, &msg.UserID, &msg.RoomID, &msg.Content, &msg.Timestamp)
	if err != nil {
		return nil, err
	}

	return &msg, nil
}
