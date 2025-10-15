package main

import (
	"net/http"

	"go-star/dal"

	"github.com/a-h/templ"
)

func (app *application) RoomHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roomList, err := dal.ListRooms(app.db)
		if err != nil {
			app.serverError(w, r, err)
			return
		}

		// Render the template with the rooms data
		templ.Handler(rooms(roomList)).ServeHTTP(w, r)
	}
}
