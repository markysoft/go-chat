package dal

// Room represents a chat room
type Room struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Chatter represents a chat user
type Chatter struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
}

// Message represents a chat message
type Message struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"userId"`
	RoomID    int64  `json:"roomId"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
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
