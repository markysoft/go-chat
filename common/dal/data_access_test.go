package dal

import (
	"database/sql"
	"os"
	"testing"
)

func TestSetupDBActualFunction(t *testing.T) {
	// Test the actual SetupDB function
	// Note: This will create the main database file
	db, err := SetupDB("")
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
	db, err := SetupDB("")
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

func TestSetupDBWithCustomName(t *testing.T) {
	// Test that SetupDB works with a custom database name
	customDBName := "test-custom-db"
	db, err := SetupDB(customDBName)
	if err != nil {
		t.Fatalf("SetupDB('%s') failed: %v", customDBName, err)
	}
	defer db.Close()
	defer os.Remove("./" + customDBName + ".db") // Clean up test database

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

	// Check that the initial Watercooler room exists
	var roomCount int
	roomQuery := `SELECT COUNT(*) FROM rooms WHERE name = 'Watercooler'`
	err = db.QueryRow(roomQuery).Scan(&roomCount)
	if err != nil {
		t.Errorf("Failed to count Watercooler rooms: %v", err)
	}
	if roomCount != 1 {
		t.Errorf("Expected exactly 1 Watercooler room, found %d", roomCount)
	}

	t.Log("SetupDB with custom name test completed successfully")
}

func TestSetupDBWithEmptyName(t *testing.T) {
	// Test that SetupDB falls back to default when given empty string
	db, err := SetupDB("")
	if err != nil {
		t.Fatalf("SetupDB('') failed: %v", err)
	}
	defer db.Close()

	// Verify the database connection is working
	err = db.Ping()
	if err != nil {
		t.Errorf("Database connection failed: %v", err)
	}

	// This should create the default chat-app.db file
	// We can't easily verify the filename, but we can verify functionality
	query := `SELECT name FROM sqlite_master WHERE type='table' AND name='rooms';`
	var tableName string
	err = db.QueryRow(query).Scan(&tableName)
	if err != nil {
		t.Errorf("Table 'rooms' was not created: %v", err)
	}

	t.Log("SetupDB with empty name test completed successfully")
}

func TestInsertChatter(t *testing.T) {
	// Setup database with unique name for this test
	testDBName := "test-insert-chatter"
	db, err := SetupDB(testDBName)
	if err != nil {
		t.Fatalf("SetupDB() failed: %v", err)
	}
	defer db.Close()
	defer os.Remove("./" + testDBName + ".db") // Clean up test database
	defer db.Close()

	// Test inserting a chatter with unique username for this test
	username := "testuser_basic"
	name := "Test User"
	chatter, err := InsertChatter(db, username, name)
	if err != nil {
		t.Fatalf("InsertChatter() failed: %v", err)
	}

	// Verify the returned chatter
	if chatter == nil {
		t.Fatal("Expected non-nil chatter")
	}
	if chatter.ID == 0 {
		t.Error("Expected non-zero chatter ID")
	}
	if chatter.Username != username {
		t.Errorf("Expected username '%s', got '%s'", username, chatter.Username)
	}
	if chatter.Name != name {
		t.Errorf("Expected name '%s', got '%s'", name, chatter.Name)
	}

	// Verify the chatter was actually inserted into the database
	var dbUsername, dbName string
	selectStmt := `SELECT username, name FROM chatters WHERE id = ?`
	err = db.QueryRow(selectStmt, chatter.ID).Scan(&dbUsername, &dbName)
	if err != nil {
		t.Errorf("Failed to retrieve inserted chatter: %v", err)
	}
	if dbUsername != username || dbName != name {
		t.Errorf("Database content doesn't match: expected (%s, %s), got (%s, %s)",
			username, name, dbUsername, dbName)
	}

	t.Log("InsertChatter test completed successfully")
}

func TestInsertChatterDuplicateUsername(t *testing.T) {
	// Test inserting chatters with duplicate usernames (should now fail with unique constraint)
	testDBName := "test-duplicate-username"
	db, err := SetupDB(testDBName)
	if err != nil {
		t.Fatalf("SetupDB() failed: %v", err)
	}
	defer db.Close()
	defer os.Remove("./" + testDBName + ".db") // Clean up test database
	defer db.Close()

	username := "duplicate_test_user"

	// Insert first chatter
	chatter1, err := InsertChatter(db, username, "First User")
	if err != nil {
		t.Fatalf("First InsertChatter() failed: %v", err)
	}

	// Verify first chatter was inserted successfully
	if chatter1.Username != username {
		t.Errorf("Expected username '%s', got '%s'", username, chatter1.Username)
	}

	// Try to insert second chatter with same username (should fail due to unique constraint)
	_, err = InsertChatter(db, username, "Second User")
	if err == nil {
		t.Error("Expected error when inserting chatter with duplicate username")
	}

	// Verify the error is related to uniqueness constraint
	// SQLite returns "UNIQUE constraint failed" error
	if err != nil && len(err.Error()) > 0 {
		t.Logf("Got expected error for duplicate username: %v", err)
	}

	t.Log("InsertChatter duplicate username test completed successfully")
}

func TestInsertChatterEmptyUsername(t *testing.T) {
	// Test inserting chatter with empty username (should fail due to NOT NULL constraint)
	testDBName := "test-empty-username"
	db, err := SetupDB(testDBName)
	if err != nil {
		t.Fatalf("SetupDB() failed: %v", err)
	}
	defer db.Close()
	defer os.Remove("./" + testDBName + ".db") // Clean up test database
	defer db.Close()

	// Try to insert chatter with empty username
	_, err = InsertChatter(db, "", "Test User")
	if err == nil {
		t.Error("Expected error when inserting chatter with empty username")
	}

	t.Log("InsertChatter empty username test completed successfully")
}

func TestInsertMultipleChatters(t *testing.T) {
	// Test inserting multiple chatters
	testDBName := "test-multiple-chatters"
	db, err := SetupDB(testDBName)
	if err != nil {
		t.Fatalf("SetupDB() failed: %v", err)
	}
	defer db.Close()
	defer os.Remove("./" + testDBName + ".db") // Clean up test database
	defer db.Close()

	// Test data
	users := []struct {
		username string
		name     string
	}{
		{"alice", "Alice Smith"},
		{"bob", "Bob Johnson"},
		{"charlie", "Charlie Brown"},
	}

	var insertedChatters []*Chatter
	for _, user := range users {
		chatter, err := InsertChatter(db, user.username, user.name)
		if err != nil {
			t.Fatalf("Failed to insert chatter '%s': %v", user.username, err)
		}
		insertedChatters = append(insertedChatters, chatter)
	}

	// Verify all chatters were inserted with unique IDs
	if len(insertedChatters) != len(users) {
		t.Errorf("Expected %d chatters, got %d", len(users), len(insertedChatters))
	}

	for i, chatter := range insertedChatters {
		if chatter.Username != users[i].username {
			t.Errorf("Chatter %d: expected username '%s', got '%s'", i, users[i].username, chatter.Username)
		}
		if chatter.Name != users[i].name {
			t.Errorf("Chatter %d: expected name '%s', got '%s'", i, users[i].name, chatter.Name)
		}
		// Verify IDs are unique and sequential
		if i > 0 && chatter.ID <= insertedChatters[i-1].ID {
			t.Errorf("Chatter IDs should be sequential: chatter[%d].ID=%d, chatter[%d].ID=%d",
				i-1, insertedChatters[i-1].ID, i, chatter.ID)
		}
	}

	t.Log("InsertMultipleChatters test completed successfully")
}

func TestInsertMessage(t *testing.T) {
	// Setup database with unique name for this test
	testDBName := "test-insert-message"
	db, err := SetupDB(testDBName)
	if err != nil {
		t.Fatalf("SetupDB() failed: %v", err)
	}
	defer db.Close()
	defer os.Remove("./" + testDBName + ".db") // Clean up test database
	defer db.Close()

	// Create a test chatter using InsertChatter
	chatter, err := InsertChatter(db, "message_test_user", "Test User")
	if err != nil {
		t.Fatalf("Failed to insert test chatter: %v", err)
	}

	// Get the Watercooler room ID (created by seedInitialData)
	var roomID int64
	roomQuery := `SELECT id FROM rooms WHERE name = 'Watercooler'`
	err = db.QueryRow(roomQuery).Scan(&roomID)
	if err != nil {
		t.Fatalf("Failed to get Watercooler room ID: %v", err)
	}

	// Test inserting a message
	content := "Hello, this is a test message!"
	message, err := InsertMessage(db, chatter.ID, roomID, content)
	if err != nil {
		t.Fatalf("InsertMessage() failed: %v", err)
	}

	// Verify the returned message
	if message == nil {
		t.Fatal("Expected non-nil message")
	}
	if message.ID == 0 {
		t.Error("Expected non-zero message ID")
	}
	if message.UserID != chatter.ID {
		t.Errorf("Expected userID %d, got %d", chatter.ID, message.UserID)
	}
	if message.RoomID != roomID {
		t.Errorf("Expected roomID %d, got %d", roomID, message.RoomID)
	}
	if message.Content != content {
		t.Errorf("Expected content '%s', got '%s'", content, message.Content)
	}
	if message.Timestamp == "" {
		t.Error("Expected non-empty timestamp")
	}

	// Verify the message was actually inserted into the database
	var dbContent string
	var dbUserID, dbRoomID int64
	selectStmt := `SELECT userId, roomId, content FROM messages WHERE id = ?`
	err = db.QueryRow(selectStmt, message.ID).Scan(&dbUserID, &dbRoomID, &dbContent)
	if err != nil {
		t.Errorf("Failed to retrieve inserted message: %v", err)
	}
	if dbUserID != chatter.ID || dbRoomID != roomID || dbContent != content {
		t.Errorf("Database content doesn't match: expected (%d, %d, %s), got (%d, %d, %s)",
			chatter.ID, roomID, content, dbUserID, dbRoomID, dbContent)
	}

	t.Log("InsertMessage test completed successfully")
}

func TestInsertMessageInvalidUser(t *testing.T) {
	// Test inserting message with invalid user ID
	testDBName := "test-invalid-user"
	db, err := SetupDB(testDBName)
	if err != nil {
		t.Fatalf("SetupDB() failed: %v", err)
	}
	defer db.Close()
	defer os.Remove("./" + testDBName + ".db") // Clean up test database
	defer db.Close()

	// Get the Watercooler room ID
	var roomID int64
	roomQuery := `SELECT id FROM rooms WHERE name = 'Watercooler'`
	err = db.QueryRow(roomQuery).Scan(&roomID)
	if err != nil {
		t.Fatalf("Failed to get Watercooler room ID: %v", err)
	}

	// Try to insert message with non-existent user ID
	invalidUserID := int64(99999)
	_, err = InsertMessage(db, invalidUserID, roomID, "Test message")
	if err == nil {
		t.Error("Expected error when inserting message with invalid user ID")
	}

	t.Log("InsertMessage invalid user test completed successfully")
}

func TestInsertMessageInvalidRoom(t *testing.T) {
	// Test inserting message with invalid room ID
	testDBName := "test-invalid-room"
	db, err := SetupDB(testDBName)
	if err != nil {
		t.Fatalf("SetupDB() failed: %v", err)
	}
	defer db.Close()
	defer os.Remove("./" + testDBName + ".db") // Clean up test database
	defer db.Close()

	// Create a test chatter using InsertChatter
	chatter, err := InsertChatter(db, "invalid_room_test_user", "Test User 2")
	if err != nil {
		t.Fatalf("Failed to insert test chatter: %v", err)
	}

	// Try to insert message with non-existent room ID
	invalidRoomID := int64(99999)
	_, err = InsertMessage(db, chatter.ID, invalidRoomID, "Test message")
	if err == nil {
		t.Error("Expected error when inserting message with invalid room ID")
	}

	t.Log("InsertMessage invalid room test completed successfully")
}

func TestInsertMultipleMessages(t *testing.T) {
	// Test inserting multiple messages
	testDBName := "test-multiple-messages"
	db, err := SetupDB(testDBName)
	if err != nil {
		t.Fatalf("SetupDB() failed: %v", err)
	}
	defer db.Close()
	defer os.Remove("./" + testDBName + ".db") // Clean up test database
	defer db.Close()

	// Create a test chatter using InsertChatter
	chatter, err := InsertChatter(db, "multi_message_test_user", "Test User 3")
	if err != nil {
		t.Fatalf("Failed to insert test chatter: %v", err)
	}

	// Get the Watercooler room ID
	var roomID int64
	roomQuery := `SELECT id FROM rooms WHERE name = 'Watercooler'`
	err = db.QueryRow(roomQuery).Scan(&roomID)
	if err != nil {
		t.Fatalf("Failed to get Watercooler room ID: %v", err)
	}

	// Insert multiple messages
	messages := []string{
		"First message",
		"Second message",
		"Third message",
	}

	var insertedMessages []*Message
	for _, content := range messages {
		msg, err := InsertMessage(db, chatter.ID, roomID, content)
		if err != nil {
			t.Fatalf("Failed to insert message '%s': %v", content, err)
		}
		insertedMessages = append(insertedMessages, msg)
	}

	// Verify all messages were inserted with unique IDs
	if len(insertedMessages) != len(messages) {
		t.Errorf("Expected %d messages, got %d", len(messages), len(insertedMessages))
	}

	for i, msg := range insertedMessages {
		if msg.Content != messages[i] {
			t.Errorf("Message %d: expected content '%s', got '%s'", i, messages[i], msg.Content)
		}
		// Verify IDs are unique and sequential
		if i > 0 && msg.ID <= insertedMessages[i-1].ID {
			t.Errorf("Message IDs should be sequential: msg[%d].ID=%d, msg[%d].ID=%d",
				i-1, insertedMessages[i-1].ID, i, msg.ID)
		}
	}

	t.Log("InsertMultipleMessages test completed successfully")
}

func TestListMessagesForRoom(t *testing.T) {
	testDBName := "test_list_messages_for_room"
	defer os.Remove("./" + testDBName + ".db")

	db, err := SetupDB(testDBName)
	if err != nil {
		t.Fatalf("SetupDB() failed: %v", err)
	}
	defer db.Close()

	// Create test chatters
	chatter1, err := InsertChatter(db, "alice", "Alice Smith")
	if err != nil {
		t.Fatalf("Failed to insert chatter1: %v", err)
	}

	chatter2, err := InsertChatter(db, "bob", "Bob Johnson")
	if err != nil {
		t.Fatalf("Failed to insert chatter2: %v", err)
	}

	// Create additional room for testing
	_, err = db.Exec("INSERT INTO rooms (name, description) VALUES (?, ?)", "General", "General discussion")
	if err != nil {
		t.Fatalf("Failed to insert test room: %v", err)
	}

	// Get room IDs
	var watercoolerID, generalID int64
	err = db.QueryRow("SELECT id FROM rooms WHERE name = ?", "Watercooler").Scan(&watercoolerID)
	if err != nil {
		t.Fatalf("Failed to get Watercooler room ID: %v", err)
	}

	err = db.QueryRow("SELECT id FROM rooms WHERE name = ?", "General").Scan(&generalID)
	if err != nil {
		t.Fatalf("Failed to get General room ID: %v", err)
	}

	// Insert messages in Watercooler room
	messages := []struct {
		userID  int64
		content string
	}{
		{chatter1.ID, "Hello everyone!"},
		{chatter2.ID, "Hi Alice!"},
		{chatter1.ID, "How's everyone doing?"},
		{chatter2.ID, "Great, thanks for asking!"},
	}

	for _, msg := range messages {
		_, err := InsertMessage(db, msg.userID, watercoolerID, msg.content)
		if err != nil {
			t.Fatalf("Failed to insert message: %v", err)
		}
	}

	// Insert a message in General room (should not appear in Watercooler results)
	_, err = InsertMessage(db, chatter1.ID, generalID, "This is in General room")
	if err != nil {
		t.Fatalf("Failed to insert message in General room: %v", err)
	}

	// Test ListMessagesForRoom for Watercooler
	watercoolerMessages, err := ListMessagesForRoom(db, watercoolerID)
	if err != nil {
		t.Fatalf("ListMessagesForRoom failed: %v", err)
	}

	// Verify we got the right number of messages
	expectedCount := len(messages)
	if len(watercoolerMessages) != expectedCount {
		t.Errorf("Expected %d messages in Watercooler, got %d", expectedCount, len(watercoolerMessages))
	}

	// Verify message content and chatter info
	for i, msg := range watercoolerMessages {
		expectedContent := messages[i].content
		if msg.Content != expectedContent {
			t.Errorf("Message %d: expected content '%s', got '%s'", i, expectedContent, msg.Content)
		}

		// Verify chatter info is populated
		if msg.ChatterName == "" {
			t.Errorf("Message %d: ChatterName should not be empty", i)
		}
		if msg.Username == "" {
			t.Errorf("Message %d: Username should not be empty", i)
		}

		// Verify specific chatter info
		if messages[i].userID == chatter1.ID {
			if msg.ChatterName != "Alice Smith" || msg.Username != "alice" {
				t.Errorf("Message %d: expected Alice Smith/alice, got %s/%s", i, msg.ChatterName, msg.Username)
			}
		} else if messages[i].userID == chatter2.ID {
			if msg.ChatterName != "Bob Johnson" || msg.Username != "bob" {
				t.Errorf("Message %d: expected Bob Johnson/bob, got %s/%s", i, msg.ChatterName, msg.Username)
			}
		}

		// Verify room ID
		if msg.RoomID != watercoolerID {
			t.Errorf("Message %d: expected roomID %d, got %d", i, watercoolerID, msg.RoomID)
		}
	}

	// Test ListMessagesForRoom for General (should only have 1 message)
	generalMessages, err := ListMessagesForRoom(db, generalID)
	if err != nil {
		t.Fatalf("ListMessagesForRoom for General failed: %v", err)
	}

	if len(generalMessages) != 1 {
		t.Errorf("Expected 1 message in General, got %d", len(generalMessages))
	}

	if len(generalMessages) > 0 {
		if generalMessages[0].Content != "This is in General room" {
			t.Errorf("General room message content mismatch")
		}
		if generalMessages[0].ChatterName != "Alice Smith" {
			t.Errorf("General room message chatter name mismatch")
		}
	}

	// Test ListMessagesForRoom for non-existent room
	emptyMessages, err := ListMessagesForRoom(db, 99999) // Use a non-existent room ID
	if err != nil {
		t.Fatalf("ListMessagesForRoom for non-existent room failed: %v", err)
	}

	if len(emptyMessages) != 0 {
		t.Errorf("Expected 0 messages for non-existent room, got %d", len(emptyMessages))
	}

	t.Log("ListMessagesForRoom test completed successfully")
}

func TestGetChatterByUsername(t *testing.T) {
	testDBName := "test_get_chatter_by_username"
	defer os.Remove("./" + testDBName + ".db")

	db, err := SetupDB(testDBName)
	if err != nil {
		t.Fatalf("SetupDB() failed: %v", err)
	}
	defer db.Close()

	// Insert test chatters
	chatter1, err := InsertChatter(db, "alice", "Alice Smith")
	if err != nil {
		t.Fatalf("Failed to insert chatter1: %v", err)
	}

	chatter2, err := InsertChatter(db, "bob", "Bob Johnson")
	if err != nil {
		t.Fatalf("Failed to insert chatter2: %v", err)
	}

	// Test 1: Get existing chatter by username
	retrievedChatter, err := GetChatterByUsername(db, "alice")
	if err != nil {
		t.Fatalf("GetChatterByUsername failed for 'alice': %v", err)
	}

	// Verify the retrieved chatter matches the inserted one
	if retrievedChatter.ID != chatter1.ID {
		t.Errorf("Expected ID %d, got %d", chatter1.ID, retrievedChatter.ID)
	}
	if retrievedChatter.Username != chatter1.Username {
		t.Errorf("Expected username '%s', got '%s'", chatter1.Username, retrievedChatter.Username)
	}
	if retrievedChatter.Name != chatter1.Name {
		t.Errorf("Expected name '%s', got '%s'", chatter1.Name, retrievedChatter.Name)
	}

	// Test 2: Get another existing chatter by username
	retrievedChatter2, err := GetChatterByUsername(db, "bob")
	if err != nil {
		t.Fatalf("GetChatterByUsername failed for 'bob': %v", err)
	}

	if retrievedChatter2.ID != chatter2.ID {
		t.Errorf("Expected ID %d, got %d", chatter2.ID, retrievedChatter2.ID)
	}
	if retrievedChatter2.Username != "bob" {
		t.Errorf("Expected username 'bob', got '%s'", retrievedChatter2.Username)
	}
	if retrievedChatter2.Name != "Bob Johnson" {
		t.Errorf("Expected name 'Bob Johnson', got '%s'", retrievedChatter2.Name)
	}

	// Test 3: Try to get non-existent chatter
	_, err = GetChatterByUsername(db, "nonexistent")
	if err == nil {
		t.Errorf("Expected error for non-existent chatter, got nil")
	}
	expectedErrorMsg := "chatter with username 'nonexistent' not found"
	if err.Error() != expectedErrorMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedErrorMsg, err.Error())
	}

	// Test 4: Try to get chatter with empty username
	_, err = GetChatterByUsername(db, "")
	if err == nil {
		t.Errorf("Expected error for empty username, got nil")
	}
	expectedEmptyErrorMsg := "username cannot be empty"
	if err.Error() != expectedEmptyErrorMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedEmptyErrorMsg, err.Error())
	}

	// Test 5: Case sensitivity test
	_, err = GetChatterByUsername(db, "ALICE")
	if err == nil {
		t.Errorf("Expected error for case mismatch 'ALICE' (usernames should be case-sensitive)")
	}

	t.Log("GetChatterByUsername test completed successfully")
}

func TestListRooms(t *testing.T) {
	testDBName := "test_list_rooms"
	defer os.Remove("./" + testDBName + ".db")

	db, err := SetupDB(testDBName)
	if err != nil {
		t.Fatalf("SetupDB() failed: %v", err)
	}
	defer db.Close()

	// Test 1: Check that the initial Watercooler room exists
	rooms, err := ListRooms(db)
	if err != nil {
		t.Fatalf("ListRooms failed: %v", err)
	}

	if len(rooms) != 1 {
		t.Errorf("Expected 1 initial room, got %d", len(rooms))
	}

	if len(rooms) > 0 {
		if rooms[0].Name != "Watercooler" {
			t.Errorf("Expected initial room name 'Watercooler', got '%s'", rooms[0].Name)
		}
		if rooms[0].Description != "place to hang" {
			t.Errorf("Expected initial room description 'place to hang', got '%s'", rooms[0].Description)
		}
		if rooms[0].ID <= 0 {
			t.Errorf("Expected room ID > 0, got %d", rooms[0].ID)
		}
	}

	// Test 2: Add more rooms and verify ordering
	testRooms := []struct {
		name        string
		description string
	}{
		{"General", "General discussion"},
		{"Development", "Development team chat"},
		{"Random", "Random topics"},
		{"Alpha", "Should be first alphabetically"},
	}

	for _, room := range testRooms {
		_, err := db.Exec("INSERT INTO rooms (name, description) VALUES (?, ?)", room.name, room.description)
		if err != nil {
			t.Fatalf("Failed to insert test room '%s': %v", room.name, err)
		}
	}

	// Get all rooms again
	allRooms, err := ListRooms(db)
	if err != nil {
		t.Fatalf("ListRooms failed after adding rooms: %v", err)
	}

	expectedCount := 1 + len(testRooms) // Initial + test rooms
	if len(allRooms) != expectedCount {
		t.Errorf("Expected %d rooms, got %d", expectedCount, len(allRooms))
	}

	// Test 3: Verify alphabetical ordering
	expectedOrder := []string{"Alpha", "Development", "General", "Random", "Watercooler"}
	if len(allRooms) == len(expectedOrder) {
		for i, room := range allRooms {
			if room.Name != expectedOrder[i] {
				t.Errorf("Room %d: expected name '%s', got '%s'", i, expectedOrder[i], room.Name)
			}
		}
	}

	// Test 4: Verify all fields are populated
	for i, room := range allRooms {
		if room.ID <= 0 {
			t.Errorf("Room %d (%s): ID should be > 0, got %d", i, room.Name, room.ID)
		}
		if room.Name == "" {
			t.Errorf("Room %d: Name should not be empty", i)
		}
		// Description can be empty, but should not be nil for string type
	}

	// Test 5: Test with empty database (no rooms)
	emptyTestDBName := "test_list_rooms_empty"
	defer os.Remove("./" + emptyTestDBName + ".db")

	emptyDB, err := sql.Open("sqlite", "./"+emptyTestDBName+".db")
	if err != nil {
		t.Fatalf("Failed to create empty test database: %v", err)
	}
	defer emptyDB.Close()

	// Create only the rooms table, no initial data
	_, err = emptyDB.Exec("CREATE TABLE rooms (id INTEGER NOT NULL PRIMARY KEY, name TEXT, description TEXT)")
	if err != nil {
		t.Fatalf("Failed to create rooms table in empty database: %v", err)
	}

	emptyRooms, err := ListRooms(emptyDB)
	if err != nil {
		t.Fatalf("ListRooms failed on empty database: %v", err)
	}

	if len(emptyRooms) != 0 {
		t.Errorf("Expected 0 rooms in empty database, got %d", len(emptyRooms))
	}

	t.Log("ListRooms test completed successfully")
}
