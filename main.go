package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/nats-io/nats.go"
)

func main() {
	const port = 3000

	// Setup NATS
	ns, nc, err := setupNATS()
	if err != nil {
		panic(err)
	}
	defer ns.Shutdown()
	defer nc.Close()

	subject := "chat-messages"

	nc.Subscribe(subject, func(msg *nats.Msg) {
		// Print message data
		data := string(msg.Data)
		fmt.Printf("Received message: %s\n", data)
	})

	// Publish a test message
	nc.Publish(subject, []byte("Hello from Go Chat!"))

	r := chi.NewRouter()
	homePage := home("Vanilla Go")

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		templ.Handler(homePage).ServeHTTP(w, r)
	})

	log.Printf("Starting server on http://localhost:%d", port)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), r); err != nil {
		panic(err)
	}
}
