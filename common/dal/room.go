package dal

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

func ListMessagesForRoom(db *sql.DB, roomName string) ([]MessageWithChatter, error) {
	query := `
		SELECT m.id, m.userId, m.roomId, m.content, m.timestamp, c.name, c.username
		FROM messages m
		JOIN chatters c ON m.userId = c.id
		JOIN rooms r ON m.roomId = r.id
		WHERE r.name = ?
		ORDER BY m.timestamp DESC`

	rows, err := db.Query(query, roomName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []MessageWithChatter
	for rows.Next() {
		var msg MessageWithChatter
		err := rows.Scan(&msg.ID, &msg.UserID, &msg.RoomID, &msg.Content, &msg.Timestamp, &msg.ChatterName, &msg.Username)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}

func ListRooms(db *sql.DB) ([]Room, error) {
	query := `SELECT id, name, description FROM rooms ORDER BY name ASC`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []Room
	for rows.Next() {
		var room Room
		err := rows.Scan(&room.ID, &room.Name, &room.Description)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return rooms, nil
}

func GetRoom(db *sql.DB, roomID int64) (*Room, error) {
	query := `SELECT id, name, description FROM rooms WHERE id = ?`

	var room Room
	err := db.QueryRow(query, roomID).Scan(&room.ID, &room.Name, &room.Description)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("room with ID %d not found", roomID)
		}
		return nil, err
	}

	return &room, nil
}
