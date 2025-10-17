package bots

import (
	"database/sql"
	"fmt"
	"go-star/common/dal"
	"log"
	"strings"

	"github.com/nats-io/nats.go"
)

type ChatBot struct {
	Name          string
	Username      string
	Id            int64
	RoomId        int64
	Vibe          string
	LastResponses []string
}

func NewPosiBot(db *sql.DB, name string, username string, roomId int64) *ChatBot {
	bot := &ChatBot{
		Name:          name,
		Username:      username,
		RoomId:        roomId,
		Vibe:          "positive",
		LastResponses: []string{},
	}
	id, err := ensureBotUserExists(db, *bot)
	if err != nil {
		log.Printf("failed to ensure bot user exists: %v", err)
		return nil
	}
	bot.Id = id
	return bot
}

func ensureBotUserExists(db *sql.DB, bot ChatBot) (int64, error) {
	chatter, _ := dal.GetChatterByUsername(db, bot.Username)
	if chatter == nil {
		var err error
		chatter, err = dal.InsertChatter(db, bot.Username, bot.Name) 
		if err != nil {
			log.Printf("failed to create new chatter: %v", err)
			return 0, err
		}
	}
	return chatter.ID, nil
}

func (bot *ChatBot) GenerateResponse(userMessage dal.MessageWithChatter) string {
	// Simple positive response generation logic
	response := "That's great to hear! Keep up the positive vibes!"

	// Store the last response
	bot.LastResponses = append(bot.LastResponses, response)
	if len(bot.LastResponses) > 5 {
		bot.LastResponses = bot.LastResponses[1:] // Keep only the last 5 responses
	}

	return response
}

func (bot *ChatBot) Listen(db *sql.DB, nc *nats.Conn) error {

	var subject = "chat-messages"
	messageChan := make(chan string, 10)

	log.Println("chatbot listening")
	// Subscribe to NATS and forward messages to the channel
	sub, err := nc.Subscribe(subject, func(msg *nats.Msg) {
		log.Println("message received from NATS")
		data := string(msg.Data)
		select {
		case messageChan <- data:
		default:
			// Channel is full, drop the message
			log.Printf("Message channel full, dropping message: %s", data)
		}
	})
	if err != nil {
		log.Printf("failed to subscribe to messages: %v", err)
		return err
	}
	defer sub.Unsubscribe()

	log.Println("chatbot subscribed")
	for message := range messageChan {

		log.Printf("Chatbot responding to message: %s", message)
		if message == "" {
			continue
		}

		// Skip if the message starts with the bot's own username
		if strings.HasPrefix(message, bot.Username+":") {
			continue
		}

		response := bot.GenerateResponse(dal.MessageWithChatter{Content: message})
		dal.InsertMessage(db, bot.Id, bot.RoomId, response)

		formattedMessage := fmt.Sprintf("%s:%s", bot.Username, response)
		if err := nc.Publish(subject, []byte(formattedMessage)); err != nil {
			log.Printf("failed to publish message: %v", err)
		}
	}
	log.Println("chatbot unsubscribed")
	return nil
}
