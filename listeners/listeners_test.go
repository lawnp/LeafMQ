package listeners

import (
	"net"
	"testing"
	"time"
)

func mockBind(conn net.Conn) {
	// For testing purposes, do nothing in this mock function
}

func TestNewTCP(t *testing.T) {
	address := "127.0.0.1"
	port := "1883"
	listener := NewTCP(address, port)

	if listener.address != address {
		t.Errorf("Expected address to be %s, but got %s", address, listener.address)
	}

	if listener.port != port {
		t.Errorf("Expected port to be %s, but got %s", port, listener.port)
	}
}

func TestTcpListener_Serve(t *testing.T) {
	address := "127.0.0.1"
	port := "1883"

	listener := NewTCP(address, port)

	// Start the TCP listener in a goroutine
	go func() {
		err := listener.Serve(mockBind)
		if err != nil {
			t.Errorf("Unexpected error while serving: %v", err)
		}
	}()

	// Wait for a short time to allow the listener to start
	time.Sleep(100 * time.Millisecond)

	// Create a client connection to the TCP listener
	conn, err := net.Dial("tcp", address+":"+port)
	if err != nil {
		t.Fatalf("Failed to establish connection: %v", err)
	}
	defer conn.Close()
	// Wait for a short time to allow the listener to close
	time.Sleep(100 * time.Millisecond)
}



