package main

import (
	"fmt"
	"go-star/common/dal"
	"go-star/routes"
	"log/slog"
	"net/http"
	"os"

	"go-star/common"
)

func main() {
	const port = 3000
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Setup NATS
	nc, cleanup, err := common.SetupNATS()
	if err != nil {
		panic(err)
	}
	defer cleanup()

	db, err := dal.SetupDB("chat-db")

	if err != nil {
		panic(err)
	}

	r := routes.Register(logger, db, nc)
	logger.Info("Starting server", "url", fmt.Sprintf("http://localhost:%d", port))

	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), r); err != nil {
		panic(err)
	}
}
