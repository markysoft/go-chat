package main

import (
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

// setupNATS creates and starts the embedded NATS server
func setupNATS() (*server.Server, *nats.Conn, error) {
	opts := &server.Options{
		Host:   "127.0.0.1",
		Port:   4222,
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
		return nil, nil, err
	}

	nc, err := nats.Connect(ns.ClientURL())
	if err != nil {
		return nil, nil, err
	}

	return ns, nc, nil
}
