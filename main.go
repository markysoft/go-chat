package main

import (
	"fmt"
	"go-star/dal"
	"go-star/handlers"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
)

func main() {
	const port = 3000
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

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

	rh := handlers.NewHandlers(logger, db, nc)
	r.Get("/", rh.RoomsPage())
	r.Get("/rooms", rh.RoomsPage())
	r.Get("/room/{id:\\d+}", rh.RoomPage())
	r.Get("/room/messages", rh.RoomMessages())
	r.Post("/room/message", rh.SendMessage())

	log.Printf("Starting server on http://localhost:%d", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), r); err != nil {
		panic(err)
	}
}
