package main

import (
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

// setupNATS creates and starts the embedded NATS server
// Returns the server, connection, cleanup function, and error
func setupNATS() (*nats.Conn, func(), error) {
	opts := &server.Options{
		Host:   "127.0.0.1",
		Port:   4223,
		NoLog:  true,
		NoSigs: true,
	}

	ns, err := server.NewServer(opts)
	if err != nil {
		return nil, nil, err
	}

	// Start the server
	go ns.Start()

	// Wait for server to be ready for connections
	if !ns.ReadyForConnections(4 * time.Second) {
		panic("NATS Server not ready in time")
	}

	nc, err := nats.Connect(ns.ClientURL())
	if err != nil {
		ns.Shutdown() // Clean up server if connection fails
		return nil, nil, err
	}

	// Return cleanup function that handles both nc.Close() and ns.Shutdown()
	cleanup := func() {
		nc.Close()
		ns.Shutdown()
	}

	return nc, cleanup, nil
}
