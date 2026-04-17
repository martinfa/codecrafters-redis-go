package main

import (
	"io"
	"net"
	"sync"
	"testing"
	"time"
)

func TestHandleWait_WithConnectedReplicasReturnsReplicaCount(t *testing.T) {
	resetReplicationStateForTest()
	defer func() {
		resetReplicationStateForTest()
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

func TestHandleWait_WithPendingWriteCommandsReturnsAcknowledgedReplicaCount(t *testing.T) {
	resetReplicationStateForTest()
	defer func() {
		resetReplicationStateForTest()
	}()

	acknowledgingMasterConnection, acknowledgingReplicaConnection := net.Pipe()
	acknowledgingMasterChannel := make(chan []byte)
	acknowledgingReplicaChannel := make(chan []byte)
	var acknowledgingMasterWaitGroup sync.WaitGroup
	var acknowledgingReplicaWaitGroup sync.WaitGroup
	acknowledgingMasterWaitGroup.Add(1)
	acknowledgingReplicaWaitGroup.Add(1)

	go listen(acknowledgingMasterConnection, acknowledgingMasterChannel)
	go eventReactor(acknowledgingMasterChannel, acknowledgingMasterConnection, &acknowledgingMasterWaitGroup, false, nil)

	processedReplicationCommandBytes := 0
	go listen(acknowledgingReplicaConnection, acknowledgingReplicaChannel)
	go eventReactor(acknowledgingReplicaChannel, acknowledgingReplicaConnection, &acknowledgingReplicaWaitGroup, true, &processedReplicationCommandBytes)

	RegisterReplica(acknowledgingMasterConnection)

	silentMasterConnection, silentReplicaConnection := net.Pipe()
	silentMasterChannel := make(chan []byte)
	var silentMasterWaitGroup sync.WaitGroup
	silentMasterWaitGroup.Add(1)

	go listen(silentMasterConnection, silentMasterChannel)
	go eventReactor(silentMasterChannel, silentMasterConnection, &silentMasterWaitGroup, false, nil)
	go func() {
		defer silentReplicaConnection.Close()
		_, _ = io.Copy(io.Discard, silentReplicaConnection)
	}()

	RegisterReplica(silentMasterConnection)

	setCommand := "*3\r\n$3\r\nSET\r\n$3\r\nfoo\r\n$3\r\n123\r\n"
	PropagateCommand([]byte(setCommand))

	response := HandleWait(&RedisCommand{
		Type: CmdWAIT,
		Args: []string{"2", "50"},
	})

	if response != ":1\r\n" {
		t.Fatalf("expected WAIT response %q, got %q", ":1\r\n", response)
	}

	if err := acknowledgingMasterConnection.Close(); err != nil {
		t.Fatalf("close acknowledging master connection: %v", err)
	}
	if err := acknowledgingReplicaConnection.Close(); err != nil {
		t.Fatalf("close acknowledging replica connection: %v", err)
	}
	if err := silentMasterConnection.Close(); err != nil {
		t.Fatalf("close silent master connection: %v", err)
	}

	waitForEventReactor := func(waitGroup *sync.WaitGroup, description string) {
		t.Helper()

		waitFinished := make(chan struct{})
		go func() {
			waitGroup.Wait()
			close(waitFinished)
		}()

		select {
		case <-waitFinished:
		case <-time.After(2 * time.Second):
			t.Fatalf("timeout waiting for %s", description)
		}
	}

	waitForEventReactor(&acknowledgingMasterWaitGroup, "acknowledging master eventReactor")
	waitForEventReactor(&acknowledgingReplicaWaitGroup, "acknowledging replica eventReactor")
	waitForEventReactor(&silentMasterWaitGroup, "silent master eventReactor")
}

func TestClientConnection_WAITWithZeroReplicasReturnsZeroImmediately(t *testing.T) {
	resetReplicationStateForTest()
	defer func() {
		resetReplicationStateForTest()
	}()

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
