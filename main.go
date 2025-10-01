package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

func main() {

	const port = 3000
	opts := &server.Options{
		Host:   "127.0.0.1",
		Port:   4222, // Default NATS port
		NoLog:  false,
		NoSigs: true,
	}

	ns, err := server.NewServer(opts)

	if err != nil {
		panic(err)
	}

	// Start the server
	go ns.Start()

	// Wait for server to be ready for connections
	if !ns.ReadyForConnections(4 * time.Second) {
		panic("not ready for connection")
	}

	nc, err := nats.Connect(ns.ClientURL())

	subject := "my-subject"

	if err != nil {
		panic(err)
	}

	nc.Subscribe(subject, func(msg *nats.Msg) {
		// Print message data
		data := string(msg.Data)
		fmt.Println(data)

		// Shutdown the server (optional)
		// ns.Shutdown()
	})

	// Publish data to the subject
	nc.Publish(subject, []byte("Hello embedded NATS!"))

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
