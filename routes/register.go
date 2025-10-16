package routes

import (
	"database/sql"
	"go-star/handlers"
	"log/slog"

	"github.com/nats-io/nats.go"

	"github.com/go-chi/chi/v5"
)

func Register(logger *slog.Logger, db *sql.DB, nc *nats.Conn) *chi.Mux {

	r := chi.NewRouter()

	rh := handlers.NewHandlers(logger, db, nc)

	r.Get("/", rh.ListRooms())
	r.Get("/room/{id:\\d+}", rh.RoomPage())
	r.Get("/room/messages", rh.ListMessages())
	r.Post("/room/message", rh.SendMessage())

	return r
}
