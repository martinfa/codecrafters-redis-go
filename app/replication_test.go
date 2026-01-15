package main

import (
	"bufio"
	"net"
	"testing"
	"time"
)

func TestInitiateHandshake_SendsPing(t *testing.T) {
	// 1. Start a mock Master server
	masterListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to start mock master: %v", err)
	}
	defer masterListener.Close()

	masterAddr := masterListener.Addr().String()
	host, port, _ := net.SplitHostPort(masterAddr)

	// 2. Prepare our replica config
	config := Config{
		IsReplica:  true,
		MasterHost: host,
		MasterPort: port,
	}

	// 3. Run the handshake in a goroutine (so it doesn't block)
	errCh := make(chan error, 1)
	go func() {
		errCh <- InitiateHandshake(config)
	}()

	// 4. Mock Master accepts the connection
	masterConn, err := masterListener.Accept()
	if err != nil {
		t.Fatalf("Mock master failed to accept connection: %v", err)
	}
	defer masterConn.Close()

	// 5. Verify the PING command was received
	reader := bufio.NewReader(masterConn)
	received, err := reader.ReadString('\n') // Read *1\r\n
	if err != nil {
		t.Fatalf("Error reading from replica: %v", err)
	}
	if received != "*1\r\n" {
		t.Errorf("Expected *1\\r\\n, got %q", received)
	}

	received, err = reader.ReadString('\n') // Read $4\r\n
	if err != nil {
		t.Fatalf("Error reading from replica: %v", err)
	}
	if received != "$4\r\n" {
		t.Errorf("Expected $4\\r\\n, got %q", received)
	}

	received, err = reader.ReadString('\n') // Read PING\r\n
	if err != nil {
		t.Fatalf("Error reading from replica: %v", err)
	}
	if received != "PING\r\n" {
		t.Errorf("Expected PING\\r\\n, got %q", received)
	}

	// Clean up
	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("InitiateHandshake returned error: %v", err)
		}
	case <-time.After(100 * time.Millisecond):
		// This is fine, the handshake might still be waiting for more steps in the future
	}
}
