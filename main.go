package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
)

func main() {
	const port = 3000

	// Setup NATS
	nc, cleanup, err := setupNATS()
	if err != nil {
		panic(err)
	}
	defer cleanup()

	subject := "chat-messages"

	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		templ.Handler(home("Vanilla Go")).ServeHTTP(w, r)
	})
	
	r.Post("/message", MessageHandler(nc, subject))
	r.Get("/messages", MessagesHandler(nc, subject))

	log.Printf("Starting server on http://localhost:%d", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), r); err != nil {
		panic(err)
	}
}
