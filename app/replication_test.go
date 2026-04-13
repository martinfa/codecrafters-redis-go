package main

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"net"
	"testing"
	"time"
)

func TestReadReplicationHandshakeRDBSnapshot_DiscardsBulkPayload(t *testing.T) {
	payload := []byte("hello-rdb")
	input := append([]byte(fmt.Sprintf("$%d\r\n", len(payload))), payload...)
	input = append(input, []byte("*1\r\n$4\r\nPING\r\n")...)
	reader := bufio.NewReader(bytes.NewReader(input))
	if err := readReplicationHandshakeRDBSnapshot(reader); err != nil {
		t.Fatalf("readReplicationHandshakeRDBSnapshot: %v", err)
	}
	rest, err := reader.ReadBytes('\n')
	if err != nil {
		t.Fatalf("read after RDB: %v", err)
	}
	if string(rest) != "*1\r\n" {
		t.Fatalf("expected next bytes to start stream after RDB, got %q", rest)
	}
}

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

	// 11. Verify PSYNC ? -1
	assertRead("*3\r\n")
	assertRead("$5\r\n")
	assertRead("PSYNC\r\n")
	assertRead("$1\r\n")
	assertRead("?\r\n")
	assertRead("$2\r\n")
	assertRead("-1\r\n")

	// 12. Master responds with +FULLRESYNC ...
	masterConn.Write([]byte("+FULLRESYNC 8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb 0\r\n"))

	// 13. Master sends RDB snapshot (same minimal empty RDB as HandlePsync)
	rdbHex := "524544495330303131fa0972656469732d76657205372e322e30fa0a72656469732d62697473c040fa056374696d65c26d08bc65fa08757365642d6d656dc2b0c41000fa08616f662d62617365c000ff10aa32556141e212"
	rdbBytes, hexDecodeError := hex.DecodeString(rdbHex)
	if hexDecodeError != nil {
		t.Fatalf("decode fixture RDB hex: %v", hexDecodeError)
	}
	if _, writeError := masterConn.Write([]byte(fmt.Sprintf("$%d\r\n%s", len(rdbBytes), string(rdbBytes)))); writeError != nil {
		t.Fatalf("write RDB bulk to replica: %v", writeError)
	}

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
