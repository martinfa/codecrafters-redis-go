package main

import (
	"bufio"
	"fmt"
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
	// Helper to check for a specific string
	assertRead := func(expected string) {
		line, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("Error reading: %v", err)
		}
		if line != expected {
			t.Errorf("Expected %q, got %q", expected, line)
		}
	}

	assertRead("*1\r\n")
	assertRead("$4\r\n")
	assertRead("PING\r\n")

	// 6. Master responds with +PONG
	masterConn.Write([]byte("+PONG\r\n"))

	// 7. Verify REPLCONF listening-port <PORT>
	assertRead("*3\r\n")
	assertRead("$8\r\n")
	assertRead("REPLCONF\r\n")
	assertRead("$14\r\n")
	assertRead("listening-port\r\n")
	portStr := fmt.Sprintf("%d", config.Port)
	assertRead(fmt.Sprintf("$%d\r\n", len(portStr)))
	assertRead(portStr + "\r\n")

	// 8. Master responds with +OK
	masterConn.Write([]byte("+OK\r\n"))

	// 9. Verify REPLCONF capa psync2
	assertRead("*3\r\n")
	assertRead("$8\r\n")
	assertRead("REPLCONF\r\n")
	assertRead("$4\r\n")
	assertRead("capa\r\n")
	assertRead("$6\r\n")
	assertRead("psync2\r\n")

	// 10. Master responds with +OK
	masterConn.Write([]byte("+OK\r\n"))

	// Clean up
	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("InitiateHandshake returned error: %v", err)
		}
	case <-time.After(500 * time.Millisecond):
		// This is fine
	}
}
