package dal

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

func SetupDB() (*sql.DB, error) {

	db, err := sql.Open("sqlite", "./chat-app.db")
	if err != nil {
		return nil, err
	}

	// Test the connection
	if err = db.Ping(); err != nil {
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
		username TEXT, 
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
