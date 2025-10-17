package dal

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

func ListMessagesForRoom(db *sql.DB, roomId int64) ([]MessageWithChatter, error) {
	query := `
		SELECT m.id, m.userId, m.roomId, m.content, m.timestamp, c.name, c.username
		FROM messages m
		JOIN chatters c ON m.userId = c.id
		WHERE m.roomId = ?
		ORDER BY m.timestamp DESC`

	rows, err := db.Query(query, roomId)
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

// InsertRoom adds a new room to the rooms table
func InsertRoom(db *sql.DB, name, description string) (*Room, error) {
	// Validate input
	if name == "" {
		return nil, fmt.Errorf("room name cannot be empty")
	}

	stmt := `INSERT INTO rooms (name, description) VALUES (?, ?)`
	result, err := db.Exec(stmt, name, description)
	if err != nil {
		return nil, err
	}

	// Get the ID of the inserted room
	roomID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Return the complete room
	room := &Room{
		ID:          roomID,
		Name:        name,
		Description: description,
	}

	return room, nil
}
