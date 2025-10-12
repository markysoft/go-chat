package main

import (
	"fmt"
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

func MessageHandler(nc *nats.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		message := &ChatItem{}
		if err := datastar.ReadSignals(r, message); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		nc.Publish(subject, []byte(message.Message))

		// Send 204 No Content - indicates success but no response body
		w.WriteHeader(http.StatusNoContent)
	}
}

// MessagesHandler handles the SSE stream for chat messages
func MessagesHandler(nc *nats.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Client connected to messages stream")
		sse := datastar.NewSSE(w, r)

		// Create a channel to receive messages from NATS
		messageChan := make(chan string, 10)

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
				err := sse.PatchElements(
					fmt.Sprintf(`<div >%s</div>`, message),
					datastar.WithSelector("#messages"),
					datastar.WithModeAppend(),
				)
				if err != nil {
					log.Printf("Failed to send message to client: %v", err)
					return
				}
			}
		}
	}
}
