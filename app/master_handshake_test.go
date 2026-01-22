package main

import (
	"bufio"
	"fmt"
	"net"
	"sync"
	"testing"
)

func TestMasterHandlesReplConf(t *testing.T) {
	// 1. Start our server (Master)
	// We'll use a random port
	port := 6381
	l, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		t.Fatalf("Failed to start listener: %v", err)
	}
	defer l.Close()

	// Run the server accept loop in a goroutine
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			clientChannel := make(chan []byte)
			go listen(conn, clientChannel)
			var wg sync.WaitGroup
			wg.Add(1)
			go eventReactor(clientChannel, conn, &wg, false)
		}
	}()

	// 2. Mock a Replica connecting to our Master
	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		t.Fatalf("Failed to connect to master: %v", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)

	// 3. Send PING
	fmt.Fprintf(conn, "*1\r\n$4\r\nPING\r\n")
	resp, _ := reader.ReadString('\n')
	if resp != "+PONG\r\n" {
		t.Errorf("Expected +PONG\r\n, got %q", resp)
	}

	// 4. Send REPLCONF listening-port
	fmt.Fprintf(conn, "*3\r\n$8\r\nREPLCONF\r\n$14\r\nlistening-port\r\n$4\r\n6380\r\n")
	resp, _ = reader.ReadString('\n')
	if resp != "+OK\r\n" {
		t.Errorf("Expected +OK\r\n, got %q", resp)
	}

	// 5. Send REPLCONF capa psync2
	fmt.Fprintf(conn, "*3\r\n$8\r\nREPLCONF\r\n$4\r\ncapa\r\n$6\r\npsync2\r\n")
	resp, _ = reader.ReadString('\n')
	if resp != "+OK\r\n" {
		t.Errorf("Expected +OK\r\n, got %q", resp)
	}
}
