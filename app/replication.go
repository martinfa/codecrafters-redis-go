package main

import (
	"bufio"
	"fmt"
	"net"
	"time"
)

// InitiateHandshake handles the replication handshake with the master server
func InitiateHandshake(config Config) error {
	if !config.IsReplica {
		return nil
	}

	masterAddress := net.JoinHostPort(config.MasterHost, config.MasterPort)
	fmt.Printf("Connecting to master at %s...\n", masterAddress)

	conn, err := net.DialTimeout("tcp", masterAddress, 5*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to master: %w", err)
	}

	reader := bufio.NewReader(conn)

	// Step 1: Send PING
	if err := sendCommand(conn, reader, "*1\r\n$4\r\nPING\r\n"); err != nil {
		return fmt.Errorf("PING failed: %w", err)
	}
	fmt.Println("Handshake: Received PONG")

	// Step 2: Send REPLCONF listening-port <PORT>
	portStr := fmt.Sprintf("%d", config.Port)
	replConfPort := fmt.Sprintf("*3\r\n$8\r\nREPLCONF\r\n$14\r\nlistening-port\r\n$%d\r\n%s\r\n", len(portStr), portStr)
	if err := sendCommand(conn, reader, replConfPort); err != nil {
		return fmt.Errorf("REPLCONF listening-port failed: %w", err)
	}
	fmt.Println("Handshake: Received OK for listening-port")

	// Step 3: Send REPLCONF capa psync2
	replConfCapa := "*3\r\n$8\r\nREPLCONF\r\n$4\r\ncapa\r\n$6\r\npsync2\r\n"
	if err := sendCommand(conn, reader, replConfCapa); err != nil {
		return fmt.Errorf("REPLCONF capa failed: %w", err)
	}
	fmt.Println("Handshake: Received OK for capa")

	return nil
}

func sendCommand(conn net.Conn, reader *bufio.Reader, command string) error {
	_, err := fmt.Fprint(conn, command)
	if err != nil {
		return err
	}

	// Wait for response (+PONG or +OK)
	response, err := reader.ReadString('\n')
	if err != nil {
		return err
	}

	// For simple strings, we just check if it's not an error (-)
	if response[0] == '-' {
		return fmt.Errorf("master returned error: %s", response)
	}

	return nil
}
