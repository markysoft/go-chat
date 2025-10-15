package common

import (
	"testing"
	"time"

	"github.com/nats-io/nats.go"
)

func TestSetupNATS(t *testing.T) {
	// Test basic NATS setup
	nc, cleanup, err := SetupNATS()
	if err != nil {
		t.Fatalf("setupNATS() failed: %v", err)
	}
	defer cleanup()

	// Verify connection is not nil
	if nc == nil {
		t.Fatal("Expected non-nil NATS connection")
	}

	// Verify connection is connected
	if !nc.IsConnected() {
		t.Error("NATS connection is not connected")
	}

	// Test that we can get server info
	if nc.ConnectedUrl() == "" {
		t.Error("Expected non-empty connected URL")
	}

	t.Log("Basic NATS setup test completed successfully")
}

func TestSetupNATSPublishSubscribe(t *testing.T) {
	// Test that we can publish and subscribe
	nc, cleanup, err := SetupNATS()
	if err != nil {
		t.Fatalf("setupNATS() failed: %v", err)
	}
	defer cleanup()

	subject := "test.subject"
	testMessage := "Hello NATS Test!"
	received := make(chan string, 1)

	// Subscribe to the test subject
	sub, err := nc.Subscribe(subject, func(msg *nats.Msg) {
		received <- string(msg.Data)
	})
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}
	defer sub.Unsubscribe()

	// Publish a message
	err = nc.Publish(subject, []byte(testMessage))
	if err != nil {
		t.Fatalf("Failed to publish: %v", err)
	}

	// Wait for the message with timeout
	select {
	case msg := <-received:
		if msg != testMessage {
			t.Errorf("Expected message '%s', got '%s'", testMessage, msg)
		}
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for message")
	}

	t.Log("Publish/Subscribe test completed successfully")
}

func TestSetupNATSMultipleConnections(t *testing.T) {
	nc1, cleanup1, err := SetupNATS()
	if err != nil {
		t.Fatalf("First setupNATS() failed: %v", err)
	}
	defer cleanup1()

	nc2, err := nats.Connect(nc1.ConnectedUrl())
	if err != nil {
		t.Fatalf("Second connection failed: %v", err)
	}
	defer nc2.Close()

	// Both connections should be valid
	if !nc1.IsConnected() {
		t.Error("First connection is not connected")
	}
	if !nc2.IsConnected() {
		t.Error("Second connection is not connected")
	}

	// Test communication between connections
	subject := "test.multi"
	testMessage := "Multi-connection test"
	received := make(chan string, 1)

	// Subscribe with first connection
	sub, err := nc1.Subscribe(subject, func(msg *nats.Msg) {
		received <- string(msg.Data)
	})
	if err != nil {
		t.Fatalf("Failed to subscribe with nc1: %v", err)
	}
	defer sub.Unsubscribe()

	// Publish with second connection
	err = nc2.Publish(subject, []byte(testMessage))
	if err != nil {
		t.Fatalf("Failed to publish with nc2: %v", err)
	}

	// Wait for the message
	select {
	case msg := <-received:
		if msg != testMessage {
			t.Errorf("Expected message '%s', got '%s'", testMessage, msg)
		}
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for message")
	}

	t.Log("Multiple connections test completed successfully")
}

func TestSetupNATSCleanup(t *testing.T) {
	// Test that cleanup function works properly
	nc, cleanup, err := SetupNATS()
	if err != nil {
		t.Fatalf("setupNATS() failed: %v", err)
	}

	// Verify connection is active
	if !nc.IsConnected() {
		t.Error("NATS connection should be connected before cleanup")
	}

	// Call cleanup
	cleanup()

	// Give it a moment to close
	time.Sleep(100 * time.Millisecond)

	// Verify connection is closed
	if nc.IsConnected() {
		t.Error("NATS connection should be closed after cleanup")
	}

	// Verify we can't publish after cleanup
	err = nc.Publish("test.cleanup", []byte("should fail"))
	if err == nil {
		t.Error("Expected error when publishing to closed connection")
	}

	t.Log("Cleanup test completed successfully")
}

func TestSetupNATSRequestReply(t *testing.T) {
	// Test request-reply pattern
	nc, cleanup, err := SetupNATS()
	if err != nil {
		t.Fatalf("setupNATS() failed: %v", err)
	}
	defer cleanup()

	subject := "test.request"

	// Set up responder
	sub, err := nc.Subscribe(subject, func(msg *nats.Msg) {
		response := "Echo: " + string(msg.Data)
		nc.Publish(msg.Reply, []byte(response))
	})
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}
	defer sub.Unsubscribe()

	// Make request
	requestMsg := "test request"
	reply, err := nc.Request(subject, []byte(requestMsg), 2*time.Second)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	expectedReply := "Echo: " + requestMsg
	if string(reply.Data) != expectedReply {
		t.Errorf("Expected reply '%s', got '%s'", expectedReply, string(reply.Data))
	}

	t.Log("Request-reply test completed successfully")
}

func TestSetupNATSConnectionStatus(t *testing.T) {
	// Test connection status and server info
	nc, cleanup, err := SetupNATS()
	if err != nil {
		t.Fatalf("setupNATS() failed: %v", err)
	}
	defer cleanup()

	// Test connection status
	if nc.Status() != nats.CONNECTED {
		t.Errorf("Expected status CONNECTED, got %v", nc.Status())
	}

	// Test that we can get stats
	stats := nc.Stats()
	// Stats should be valid (uint64 values are always >= 0)
	if stats.InMsgs == 0 && stats.OutMsgs == 0 {
		// This is fine - no messages yet
		t.Log("No messages in stats yet (normal for fresh connection)")
	}

	// Test server info
	serverInfo := nc.ConnectedServerName()
	if serverInfo == "" {
		t.Error("Expected non-empty server name")
	}

	t.Log("Connection status test completed successfully")
}
