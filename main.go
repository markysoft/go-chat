package main

import (
	"fmt"
	"go-star/dal"
	"go-star/routes"
	"log"
	"log/slog"
	"net/http"
	"os"
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

	r := routes.Register(logger, db, nc)

	log.Printf("Starting server on http://localhost:%d", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), r); err != nil {
		panic(err)
	}
}
