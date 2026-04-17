package main

import (
	"io"
	"net"
	"sync"
	"testing"
	"time"
)

func TestHandleWait_WithConnectedReplicasReturnsReplicaCount(t *testing.T) {
	replicasMutex.Lock()
	replicas = nil
	replicasMutex.Unlock()
	defer func() {
		replicasMutex.Lock()
		replicas = nil
		replicasMutex.Unlock()
	}()

	firstReplicaConnection, firstReplicaPeer := net.Pipe()
	defer firstReplicaConnection.Close()
	defer firstReplicaPeer.Close()

	secondReplicaConnection, secondReplicaPeer := net.Pipe()
	defer secondReplicaConnection.Close()
	defer secondReplicaPeer.Close()

	RegisterReplica(firstReplicaConnection)
	RegisterReplica(secondReplicaConnection)

	command := &RedisCommand{
		Type: CmdWAIT,
		Args: []string{"9", "500"},
	}

	response := HandleWait(command)

	if response != ":2\r\n" {
		t.Fatalf("expected WAIT response %q, got %q", ":2\r\n", response)
	}
}

func TestClientConnection_WAITWithZeroReplicasReturnsZeroImmediately(t *testing.T) {
	clientConnection, serverConnection := net.Pipe()
	defer clientConnection.Close()

	commandChannel := make(chan []byte)
	var waitGroup sync.WaitGroup
	waitGroup.Add(1)

	go listen(serverConnection, commandChannel)
	go eventReactor(commandChannel, serverConnection, &waitGroup, false, nil)

	waitCommand := "*3\r\n$4\r\nWAIT\r\n$1\r\n0\r\n$5\r\n60000\r\n"
	if _, writeError := clientConnection.Write([]byte(waitCommand)); writeError != nil {
		t.Fatalf("write WAIT command: %v", writeError)
	}

	if err := clientConnection.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatalf("set read deadline: %v", err)
	}

	responseBuffer := make([]byte, len(":0\r\n"))
	if _, readError := io.ReadFull(clientConnection, responseBuffer); readError != nil {
		t.Fatalf("read WAIT response: %v", readError)
	}

	if string(responseBuffer) != ":0\r\n" {
		t.Fatalf("expected WAIT response %q, got %q", ":0\r\n", string(responseBuffer))
	}

	if err := clientConnection.Close(); err != nil {
		t.Fatalf("close client side: %v", err)
	}

	waitFinished := make(chan struct{})
	go func() {
		waitGroup.Wait()
		close(waitFinished)
	}()

	select {
	case <-waitFinished:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for eventReactor to finish after client close")
	}
}
