package main

import (
	"fmt"
	"go-star/dal"
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

	db, err := dal.SetupDB("chat-db")
	if err != nil {
		panic(err)
	}

	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		templ.Handler(home("Vanilla Go")).ServeHTTP(w, r)
	})

	r.Post("/message", MessageHandler(nc, db))
	r.Get("/messages", MessagesHandler(nc, db))

	log.Printf("Starting server on http://localhost:%d", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), r); err != nil {
		panic(err)
	}
}
