package main

import (
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

	// Use a timeout for the connection
	conn, err := net.DialTimeout("tcp", masterAddress, 5*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to master: %w", err)
	}
	// We do NOT close the connection here because we'll need it for the rest of the handshake
	// and to receive the replication stream later.

	// Step 1: Send PING
	// Format: *1\r\n$4\r\nPING\r\n
	_, err = fmt.Fprintf(conn, "*1\r\n$4\r\nPING\r\n")
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to send PING to master: %w", err)
	}

	fmt.Println("Sent PING to master")

	return nil
}
