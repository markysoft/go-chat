package main

import (
	"fmt"
	"go-star/dal"
	"log"
	"log/slog"
	"net/http"
	"os"
	"database/sql"
	"github.com/nats-io/nats.go"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
)

type application struct {
	logger *slog.Logger
	db		 *sql.DB
	nc		 *nats.Conn
}

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
	app := &application{
		logger: logger,
		db:     db,
		nc:     nc,
	}

	if err != nil {
		panic(err)
	}

	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		templ.Handler(home("Vanilla Go")).ServeHTTP(w, r)
	})

	r.Post("/message", app.MessageHandler())
	r.Get("/messages", app.MessagesHandler())

	log.Printf("Starting server on http://localhost:%d", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), r); err != nil {
		panic(err)
	}
}
