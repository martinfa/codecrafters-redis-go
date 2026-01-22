package main

import (
	"bufio"
	"net"
	"testing"
	"time"
)

func TestCommandPropagation(t *testing.T) {
	// 1. Reset replicas for the test
	replicasMutex.Lock()
	replicas = nil
	replicasMutex.Unlock()

	// 2. Start a mock replica server
	replicaListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to start mock replica: %v", err)
	}
	defer replicaListener.Close()

	replicaAddr := replicaListener.Addr().String()

	// 3. Connect to the mock replica from the "Master" side
	// In the real app, the replica connects to the master.
	// Here we simulate the master holding the connection.
	masterToReplicaConn, err := net.Dial("tcp", replicaAddr)
	if err != nil {
		t.Fatalf("Failed to connect to mock replica: %v", err)
	}
	defer masterToReplicaConn.Close()

	// 4. Register the replica in our master's registry
	RegisterReplica(masterToReplicaConn)

	// 5. Accept the connection on the replica side
	replicaConn, err := replicaListener.Accept()
	if err != nil {
		t.Fatalf("Mock replica failed to accept connection: %v", err)
	}
	defer replicaConn.Close()

	// 6. Propagate a command
	testCommand := "*3\r\n$3\r\nSET\r\n$3\r\nfoo\r\n$3\r\nbar\r\n"
	PropagateCommand([]byte(testCommand))

	// 7. Verify the replica received the command
	reader := bufio.NewReader(replicaConn)
	received := make([]byte, len(testCommand))

	// Set a deadline so the test doesn't hang if it fails
	replicaConn.SetReadDeadline(time.Now().Add(1 * time.Second))

	_, err = reader.Read(received)
	if err != nil {
		t.Fatalf("Error reading propagated command: %v", err)
	}

	if string(received) != testCommand {
		t.Errorf("Expected propagated command %q, got %q", testCommand, string(received))
	}
}

func TestIsWriteCommand(t *testing.T) {
	tests := []struct {
		cmdType  CommandType
		expected bool
	}{
		{CmdSET, true},
		{CmdGET, false},
		{CmdPING, false},
		{CmdECHO, false},
		{CmdINFO, false},
	}

	for _, tt := range tests {
		t.Run(tt.cmdType.String(), func(t *testing.T) {
			if tt.cmdType.IsWrite() != tt.expected {
				t.Errorf("CommandType %s.IsWrite() = %v, expected %v", tt.cmdType.String(), tt.cmdType.IsWrite(), tt.expected)
			}
		})
	}
}
