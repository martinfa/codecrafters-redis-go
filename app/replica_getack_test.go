package main

import (
	"bytes"
	"net"
	"sync"
	"testing"
	"time"
)

func TestReplicaMasterConnection_ReplconfGetackFromMaster_YieldsAckWithOffsetZero(t *testing.T) {
	replicaConnection, masterConnection := net.Pipe()
	defer masterConnection.Close()

	commandChannel := make(chan []byte)
	var waitGroup sync.WaitGroup
	waitGroup.Add(1)

	processedReplicationCommandBytes := 0
	go listen(replicaConnection, commandChannel)
	go eventReactor(commandChannel, replicaConnection, &waitGroup, true, &processedReplicationCommandBytes)

	masterGetackCommand := "*3\r\n$8\r\nreplconf\r\n$6\r\ngetack\r\n$1\r\n*\r\n"
	if _, err := masterConnection.Write([]byte(masterGetackCommand)); err != nil {
		t.Fatalf("write GETACK to pipe: %v", err)
	}

	if err := masterConnection.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatalf("set read deadline: %v", err)
	}
	responseBuffer := make([]byte, 256)
	responseLength, err := masterConnection.Read(responseBuffer)
	if err != nil {
		t.Fatalf("read ACK from replica: %v", err)
	}
	actualResponse := responseBuffer[:responseLength]
	expectedResponse := []byte(FormatReplicaReplconfAcknowledgementRESPArray(0))
	if !bytes.Equal(actualResponse, expectedResponse) {
		t.Fatalf("expected ACK %q, got %q", expectedResponse, actualResponse)
	}

	if err := masterConnection.Close(); err != nil {
		t.Fatalf("close master side: %v", err)
	}

	waitFinished := make(chan struct{})
	go func() {
		waitGroup.Wait()
		close(waitFinished)
	}()
	select {
	case <-waitFinished:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for eventReactor to finish after connection close")
	}
}
