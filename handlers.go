package main

import (
	"fmt"
	"go-star/dal"
	"log"
	"net/http"

	"github.com/nats-io/nats.go"
	"github.com/starfederation/datastar-go/datastar"
)

var subject = "chat-messages"

type ChatItem struct {
	Message  string `json:"message"`
	Username string `json:"username"`
}

func (app *application) MessageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		message := &ChatItem{}
		if err := datastar.ReadSignals(r, message); err != nil {
			app.serverError(w, r, err)
			return
		}

		userID, err := getUserID(w, r)
		if err != nil {
			app.serverError(w, r, fmt.Errorf("failed to initialize user session: %w", err))
			return
		}

		chatter, _ := dal.GetChatterByUsername(app.db, userID)
		if chatter == nil {
			chatter, _ = dal.InsertChatter(app.db, userID, "Some User")
		}

		_, err = dal.InsertMessage(app.db, chatter.ID, 1, message.Message)
		if err != nil {
			app.serverError(w, r, fmt.Errorf("failed to save message: %w", err))
			return
		}
		app.nc.Publish(subject, []byte(message.Message))

		// Send 204 No Content - indicates success but no response body
		w.WriteHeader(http.StatusNoContent)
	}
}

// MessagesHandler handles the SSE stream for chat messages
func (app *application) MessagesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check for userId cookie or generate a new one
		userID, err := getUserID(w, r)
		if err != nil {
			app.serverError(w, r, fmt.Errorf("failed to get or generate user ID: %w", err))
			return
		}

		log.Printf("Client connected to messages stream with userID: %s", userID)
		sse := datastar.NewSSE(w, r)

		// Create a channel to receive messages from NATS
		messageChan := make(chan string, 10)

		// Subscribe to NATS and forward messages to the channel
		sub, err := app.nc.Subscribe(subject, func(msg *nats.Msg) {
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
			http.Error(w, "Failed to subscribe to messages", http.StatusInternalServerError)
			return
		}
		defer sub.Unsubscribe()

		// Keep the connection alive and listen for messages
		for {
			select {
			case <-r.Context().Done():
				// Client disconnected
				log.Println("Client disconnected from messages stream")
				return
			case message := <-messageChan:
				log.Printf("Sending message to client: %s", message)
				if message == "" {
					continue
				}
				allMessages, err := dal.ListMessagesForRoom(app.db, "Watercooler")
				if err != nil {
					log.Printf("Failed to list messages: %v", err)
					continue
				}
				err = sse.PatchElementTempl(messages(allMessages))
				if err != nil {
					log.Printf("Failed to send message to client: %v", err)
					return
				}
			}
		}
	}
}
