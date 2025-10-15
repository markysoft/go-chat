package dal

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

func SetupDB(dbName string) (*sql.DB, error) {
	// Default database name if empty
	if dbName == "" {
		dbName = "chat-app"
	}

	dbPath := fmt.Sprintf("./%s.db", dbName)
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	// Test the connection
	if err = db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	// Enable foreign key constraints
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		db.Close()
		return nil, err
	}

	// Create all tables
	createFuncs := []func(*sql.DB) error{
		createRooms,
		createChatters,
		createMessages,
	}

	for _, createFunc := range createFuncs {
		if err := createFunc(db); err != nil {
			db.Close()
			return nil, err
		}
	}

	// Seed initial data
	if err := seedInitialData(db); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

// createTable executes a CREATE TABLE statement
func createTable(db *sql.DB, tableName, schema string) error {
	stmt := "CREATE TABLE IF NOT EXISTS " + tableName + " (" + schema + ");"
	_, err := db.Exec(stmt)
	return err
}

// Table schemas
const (
	roomsSchema = `
		id INTEGER NOT NULL PRIMARY KEY, 
		name TEXT, 
		description TEXT`

	chattersSchema = `
		id INTEGER NOT NULL PRIMARY KEY, 
		username TEXT UNIQUE NOT NULL,
		name TEXT`

	messagesSchema = `
		id INTEGER NOT NULL PRIMARY KEY, 
		userId INTEGER NOT NULL, 
		roomId INTEGER NOT NULL,
		content TEXT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(userId) REFERENCES chatters(id),
		FOREIGN KEY(roomId) REFERENCES rooms(id)`
)

// seedInitialData adds default data if it doesn't exist
func seedInitialData(db *sql.DB) error {
	// Check if Watercooler room already exists
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM rooms WHERE name = ?", "Watercooler").Scan(&count)
	if err != nil {
		return err
	}

	// If Watercooler doesn't exist, create it
	if count == 0 {
		_, err = db.Exec("INSERT INTO rooms (name, description) VALUES (?, ?)", "Watercooler", "place to hang")
		if err != nil {
			return err
		}
	}

	return nil
}

func createRooms(db *sql.DB) error {
	return createTable(db, "rooms", roomsSchema)
}

func createChatters(db *sql.DB) error {
	return createTable(db, "chatters", chattersSchema)
}

func createMessages(db *sql.DB) error {
	return createTable(db, "messages", messagesSchema)
}

// Chatter represents a chat user
type Chatter struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
}

// InsertChatter adds a new chatter to the chatters table
func InsertChatter(db *sql.DB, username, name string) (*Chatter, error) {
	// Validate input
	if username == "" {
		return nil, fmt.Errorf("username cannot be empty")
	}

	stmt := `INSERT INTO chatters (username, name) VALUES (?, ?)`
	result, err := db.Exec(stmt, username, name)
	if err != nil {
		return nil, err
	}

	// Get the ID of the inserted chatter
	chatterID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Return the complete chatter
	chatter := &Chatter{
		ID:       chatterID,
		Username: username,
		Name:     name,
	}

	return chatter, nil
}

// GetChatterByUsername retrieves a chatter by their username
func GetChatterByUsername(db *sql.DB, username string) (*Chatter, error) {
	// Validate input
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

// Message represents a chat message
type Message struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"userId"`
	RoomID    int64  `json:"roomId"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
}

// InsertMessage adds a new message to the messages table
func InsertMessage(db *sql.DB, userID, roomID int64, content string) (*Message, error) {
	stmt := `INSERT INTO messages (userId, roomId, content) VALUES (?, ?, ?)`
	result, err := db.Exec(stmt, userID, roomID, content)
	if err != nil {
		return nil, err
	}

	// Get the ID of the inserted message
	messageID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Retrieve the complete message with timestamp
	selectStmt := `SELECT id, userId, roomId, content, timestamp FROM messages WHERE id = ?`
	var msg Message
	err = db.QueryRow(selectStmt, messageID).Scan(&msg.ID, &msg.UserID, &msg.RoomID, &msg.Content, &msg.Timestamp)
	if err != nil {
		return nil, err
	}

	return &msg, nil
}

// MessageWithChatter represents a message with the chatter's name included
type MessageWithChatter struct {
	ID          int64  `json:"id"`
	UserID      int64  `json:"userId"`
	RoomID      int64  `json:"roomId"`
	Content     string `json:"content"`
	Timestamp   string `json:"timestamp"`
	ChatterName string `json:"chatterName"`
	Username    string `json:"username"`
}

// ListMessagesForRoom retrieves all messages for a given room name with chatter info
func ListMessagesForRoom(db *sql.DB, roomName string) ([]MessageWithChatter, error) {
	query := `
		SELECT m.id, m.userId, m.roomId, m.content, m.timestamp, c.name, c.username
		FROM messages m
		JOIN chatters c ON m.userId = c.id
		JOIN rooms r ON m.roomId = r.id
		WHERE r.name = ?
		ORDER BY m.timestamp ASC`

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

// Room represents a chat room
type Room struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ListRooms retrieves all rooms from the database
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
