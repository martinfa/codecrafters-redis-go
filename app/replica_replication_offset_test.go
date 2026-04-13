package main

import (
	"bytes"
	"io"
	"net"
	"sync"
	"testing"
	"time"
)

func TestFormatReplicaReplconfAcknowledgementRESPArray(t *testing.T) {
	tests := []struct {
		name                           string
		processedReplicationByteOffset int
		expectedRESP                   string
	}{
		{
			name:                           "zero offset",
			processedReplicationByteOffset: 0,
			expectedRESP:                   "*3\r\n$8\r\nREPLCONF\r\n$3\r\nACK\r\n$1\r\n0\r\n",
		},
		{
			name:                           "two digit offset",
			processedReplicationByteOffset: 51,
			expectedRESP:                   "*3\r\n$8\r\nREPLCONF\r\n$3\r\nACK\r\n$2\r\n51\r\n",
		},
		{
			name:                           "three digit offset",
			processedReplicationByteOffset: 146,
			expectedRESP:                   "*3\r\n$8\r\nREPLCONF\r\n$3\r\nACK\r\n$3\r\n146\r\n",
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			actual := FormatReplicaReplconfAcknowledgementRESPArray(testCase.processedReplicationByteOffset)
			if actual != testCase.expectedRESP {
				t.Fatalf("expected %q, got %q", testCase.expectedRESP, actual)
			}
		})
	}
}

func TestReplicaMasterConnection_ReplicationOffsetAcrossGetackPingGetack(t *testing.T) {
	replicaConnection, masterConnection := net.Pipe()
	defer masterConnection.Close()

	masterReplconfGetackCommand := "*3\r\n$8\r\nreplconf\r\n$6\r\ngetack\r\n$1\r\n*\r\n"
	masterPingCommand := "*1\r\n$4\r\nPING\r\n"

	commandChannel := make(chan []byte)
	var waitGroup sync.WaitGroup
	waitGroup.Add(1)

	processedReplicationCommandBytes := 0
	go listen(replicaConnection, commandChannel)
	go eventReactor(commandChannel, replicaConnection, &waitGroup, true, &processedReplicationCommandBytes)

	readDeadlineError := masterConnection.SetReadDeadline(time.Now().Add(3 * time.Second))
	if readDeadlineError != nil {
		t.Fatalf("set read deadline: %v", readDeadlineError)
	}

	firstAckExpected := FormatReplicaReplconfAcknowledgementRESPArray(0)
	if _, writeError := masterConnection.Write([]byte(masterReplconfGetackCommand)); writeError != nil {
		t.Fatalf("write first GETACK: %v", writeError)
	}
	firstAckBuffer := make([]byte, len(firstAckExpected))
	if _, readError := io.ReadFull(masterConnection, firstAckBuffer); readError != nil {
		t.Fatalf("read first ACK: %v", readError)
	}
	if !bytes.Equal(firstAckBuffer, []byte(firstAckExpected)) {
		t.Fatalf("first ACK: expected %q, got %q", firstAckExpected, firstAckBuffer)
	}

	getackByteLength := len(masterReplconfGetackCommand)
	secondAckExpectedOffset := getackByteLength + len(masterPingCommand)
	if _, writeError := masterConnection.Write([]byte(masterPingCommand)); writeError != nil {
		t.Fatalf("write PING: %v", writeError)
	}
	if _, writeError := masterConnection.Write([]byte(masterReplconfGetackCommand)); writeError != nil {
		t.Fatalf("write second GETACK: %v", writeError)
	}

	secondAckExpected := FormatReplicaReplconfAcknowledgementRESPArray(secondAckExpectedOffset)
	secondAckBuffer := make([]byte, len(secondAckExpected))
	if _, readError := io.ReadFull(masterConnection, secondAckBuffer); readError != nil {
		t.Fatalf("read second ACK: %v", readError)
	}
	if !bytes.Equal(secondAckBuffer, []byte(secondAckExpected)) {
		t.Fatalf("second ACK: expected %q (offset before second GETACK = %d), got %q",
			secondAckExpected, secondAckExpectedOffset, secondAckBuffer)
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
		t.Fatal("timeout waiting for eventReactor")
	}
}
