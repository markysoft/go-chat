package dal

import (
	"testing"
)

func TestSetupDBActualFunction(t *testing.T) {
	// Test the actual SetupDB function
	// Note: This will create the main database file
	db, err := SetupDB()
	if err != nil {
		t.Fatalf("SetupDB() failed: %v", err)
	}
	defer db.Close()

	// Verify the database connection is working
	err = db.Ping()
	if err != nil {
		t.Errorf("Database connection failed: %v", err)
	}

	// Check that the rooms table was created
	query := `SELECT name FROM sqlite_master WHERE type='table' AND name='rooms';`
	var tableName string
	err = db.QueryRow(query).Scan(&tableName)
	if err != nil {
		t.Errorf("Table 'rooms' was not created: %v", err)
	}
	if tableName != "rooms" {
		t.Errorf("Expected table name 'rooms', got '%s'", tableName)
	}

	t.Log("SetupDB function test completed successfully")
}

func TestSetupDBCreatesInitialRoom(t *testing.T) {
	// Test that SetupDB creates the initial Watercooler room
	db, err := SetupDB()
	if err != nil {
		t.Fatalf("SetupDB() failed: %v", err)
	}
	defer db.Close()

	// Check that the Watercooler room exists
	var name, description string
	query := `SELECT name, description FROM rooms WHERE name = 'Watercooler'`
	err = db.QueryRow(query).Scan(&name, &description)
	if err != nil {
		t.Errorf("Initial Watercooler room was not created: %v", err)
		return
	}

	// Verify the room details
	if name != "Watercooler" {
		t.Errorf("Expected room name 'Watercooler', got '%s'", name)
	}
	if description != "place to hang" {
		t.Errorf("Expected description 'place to hang', got '%s'", description)
	}

	// Verify only one Watercooler room exists (no duplicates)
	var count int
	countQuery := `SELECT COUNT(*) FROM rooms WHERE name = 'Watercooler'`
	err = db.QueryRow(countQuery).Scan(&count)
	if err != nil {
		t.Errorf("Failed to count Watercooler rooms: %v", err)
		return
	}
	if count != 1 {
		t.Errorf("Expected exactly 1 Watercooler room, found %d", count)
	}

	t.Log("Initial room creation test completed successfully")
}
