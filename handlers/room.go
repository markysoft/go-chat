package handlers

import (
	"net/http"
	"database/sql"
	"go-star/dal"
	"go-star/handlers/pages"
	"log/slog"

	"github.com/a-h/templ"
	"github.com/nats-io/nats.go"
)

type Handlers struct {
		logger *slog.Logger
	db     *sql.DB
	nc     *nats.Conn
}

func NewHandlers(logger *slog.Logger, db *sql.DB, nc *nats.Conn) *Handlers {
	return &Handlers{
		logger: logger,
		db:     db,
		nc:     nc,
	}
}

func (h *Handlers) RoomPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roomList, err := dal.ListRooms(h.db)
		if err != nil {
			h.logger.Error("Error listing rooms", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Render the template with the rooms data
		templ.Handler(pages.RoomsList(roomList)).ServeHTTP(w, r)
	}
}
