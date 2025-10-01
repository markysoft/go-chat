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
