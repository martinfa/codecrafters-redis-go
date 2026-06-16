package main

import (
	"net"
	"testing"
	"time"
)

func resetPublishTestState(t *testing.T) {
	t.Helper()
	ResetConnectionPubSubStatesForTest()
	ResetConnectionWriteMutexesForTest()
}

func testConnectionPair(t *testing.T) (net.Conn, net.Conn) {
	t.Helper()

	serverConnection, clientConnection := net.Pipe()
	t.Cleanup(func() {
		serverConnection.Close()
		clientConnection.Close()
	})

	return serverConnection, clientConnection
}

func startDrainConnection(connection net.Conn) {
	go func() {
		buffer := make([]byte, 4096)
		for {
			_, readError := connection.Read(buffer)
			if readError != nil {
				return
			}
		}
	}()
}

func readAvailableBytes(connection net.Conn) ([]byte, error) {
	buffer := make([]byte, 4096)
	bytesRead, readError := connection.Read(buffer)
	if readError != nil {
		return nil, readError
	}

	return buffer[:bytesRead], nil
}

func formatExpectedPubSubMessageResponse(channel string, message string) string {
	return encodePubSubMessageResponse(channel, message)
}

func TestParsePublishCommandArguments(t *testing.T) {
	tests := []struct {
		name            string
		args            []string
		expectedChannel string
		expectedMessage string
		expectedError   string
	}{
		{
			name:            "parses channel and message",
			args:            []string{"bar", "msg"},
			expectedChannel: "bar",
			expectedMessage: "msg",
		},
		{
			name:          "rejects no arguments",
			args:          []string{},
			expectedError: "-ERR wrong number of arguments for 'publish' command\r\n",
		},
		{
			name:          "rejects missing message",
			args:          []string{"bar"},
			expectedError: "-ERR wrong number of arguments for 'publish' command\r\n",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			channel, message, errorResponse := parsePublishCommandArguments(&RedisCommand{
				Type: CmdPUBLISH,
				Args: testCase.args,
			})

			if errorResponse != testCase.expectedError {
				t.Errorf("parsePublishCommandArguments() error = %q, expected %q", errorResponse, testCase.expectedError)
			}

			if channel != testCase.expectedChannel {
				t.Errorf("parsePublishCommandArguments() channel = %q, expected %q", channel, testCase.expectedChannel)
			}

			if message != testCase.expectedMessage {
				t.Errorf("parsePublishCommandArguments() message = %q, expected %q", message, testCase.expectedMessage)
			}
		})
	}
}

func TestHandlePublishReturnsZeroWhenNoSubscribers(t *testing.T) {
	resetPublishTestState(t)

	result := HandlePublish(&RedisCommand{
		Type: CmdPUBLISH,
		Args: []string{"bar", "msg"},
	})

	if result != ":0\r\n" {
		t.Errorf("HandlePublish() = %q, expected %q", result, ":0\r\n")
	}
}

func TestHandlePublishCodecraftersStageExample(t *testing.T) {
	resetPublishTestState(t)

	firstServerConnection, firstClientConnection := testConnectionPair(t)
	secondServerConnection, secondClientConnection := testConnectionPair(t)
	thirdServerConnection, thirdClientConnection := testConnectionPair(t)
	startDrainConnection(firstClientConnection)
	startDrainConnection(secondClientConnection)
	startDrainConnection(thirdClientConnection)

	HandleSubscribe(firstServerConnection, &RedisCommand{
		Type: CmdSUBSCRIBE,
		Args: []string{"foo"},
	})
	HandleSubscribe(secondServerConnection, &RedisCommand{
		Type: CmdSUBSCRIBE,
		Args: []string{"bar"},
	})
	HandleSubscribe(thirdServerConnection, &RedisCommand{
		Type: CmdSUBSCRIBE,
		Args: []string{"bar"},
	})

	barResult := HandlePublish(&RedisCommand{
		Type: CmdPUBLISH,
		Args: []string{"bar", "msg"},
	})
	if barResult != ":2\r\n" {
		t.Errorf("HandlePublish(bar) = %q, expected %q", barResult, ":2\r\n")
	}

	fooResult := HandlePublish(&RedisCommand{
		Type: CmdPUBLISH,
		Args: []string{"foo", "msg"},
	})
	if fooResult != ":1\r\n" {
		t.Errorf("HandlePublish(foo) = %q, expected %q", fooResult, ":1\r\n")
	}
}

func TestHandlePublishCountsOneClientSubscribedToMultipleChannelsOncePerChannel(t *testing.T) {
	resetPublishTestState(t)

	serverConnection, clientConnection := testConnectionPair(t)
	startDrainConnection(clientConnection)

	HandleSubscribe(serverConnection, &RedisCommand{
		Type: CmdSUBSCRIBE,
		Args: []string{"foo"},
	})
	HandleSubscribe(serverConnection, &RedisCommand{
		Type: CmdSUBSCRIBE,
		Args: []string{"bar"},
	})

	fooResult := HandlePublish(&RedisCommand{
		Type: CmdPUBLISH,
		Args: []string{"foo", "msg"},
	})
	if fooResult != ":1\r\n" {
		t.Errorf("HandlePublish(foo) = %q, expected %q", fooResult, ":1\r\n")
	}

	barResult := HandlePublish(&RedisCommand{
		Type: CmdPUBLISH,
		Args: []string{"bar", "msg"},
	})
	if barResult != ":1\r\n" {
		t.Errorf("HandlePublish(bar) = %q, expected %q", barResult, ":1\r\n")
	}
}

func TestHandlePublishDeliversMessageToSubscribedClients(t *testing.T) {
	resetPublishTestState(t)

	firstServerConnection, firstClientConnection := testConnectionPair(t)
	secondServerConnection, secondClientConnection := testConnectionPair(t)
	thirdServerConnection, thirdClientConnection := testConnectionPair(t)

	HandleSubscribe(firstServerConnection, &RedisCommand{
		Type: CmdSUBSCRIBE,
		Args: []string{"foo"},
	})
	HandleSubscribe(secondServerConnection, &RedisCommand{
		Type: CmdSUBSCRIBE,
		Args: []string{"foo"},
	})
	HandleSubscribe(thirdServerConnection, &RedisCommand{
		Type: CmdSUBSCRIBE,
		Args: []string{"bar"},
	})

	expectedMessage := formatExpectedPubSubMessageResponse("foo", "hello")

	resultChannel := make(chan string, 1)
	go func() {
		resultChannel <- HandlePublish(&RedisCommand{
			Type: CmdPUBLISH,
			Args: []string{"foo", "hello"},
		})
	}()

	firstClientMessage, readError := readAvailableBytes(firstClientConnection)
	if readError != nil {
		t.Fatalf("read first client message: %v", readError)
	}
	if string(firstClientMessage) != expectedMessage {
		t.Errorf("first client message = %q, expected %q", string(firstClientMessage), expectedMessage)
	}

	secondClientMessage, readError := readAvailableBytes(secondClientConnection)
	if readError != nil {
		t.Fatalf("read second client message: %v", readError)
	}
	if string(secondClientMessage) != expectedMessage {
		t.Errorf("second client message = %q, expected %q", string(secondClientMessage), expectedMessage)
	}

	select {
	case publishResult := <-resultChannel:
		if publishResult != ":2\r\n" {
			t.Fatalf("HandlePublish(foo) = %q, expected %q", publishResult, ":2\r\n")
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for publish result")
	}

	thirdClientConnection.SetReadDeadline(time.Now().Add(50 * time.Millisecond))
	_, readError = readAvailableBytes(thirdClientConnection)
	if readError == nil {
		t.Fatal("third client should not receive message for foo channel")
	}
}

func TestHandlePublishDeliversMessageOnlyToMatchingChannelSubscribers(t *testing.T) {
	resetPublishTestState(t)

	fooServerConnection, fooClientConnection := testConnectionPair(t)
	barServerConnection, barClientConnection := testConnectionPair(t)

	HandleSubscribe(fooServerConnection, &RedisCommand{
		Type: CmdSUBSCRIBE,
		Args: []string{"foo"},
	})
	HandleSubscribe(barServerConnection, &RedisCommand{
		Type: CmdSUBSCRIBE,
		Args: []string{"bar"},
	})

	expectedBarMessage := formatExpectedPubSubMessageResponse("bar", "world")

	resultChannel := make(chan string, 1)
	go func() {
		resultChannel <- HandlePublish(&RedisCommand{
			Type: CmdPUBLISH,
			Args: []string{"bar", "world"},
		})
	}()

	barClientMessage, readError := readAvailableBytes(barClientConnection)
	if readError != nil {
		t.Fatalf("read bar client message: %v", readError)
	}
	if string(barClientMessage) != expectedBarMessage {
		t.Errorf("bar client message = %q, expected %q", string(barClientMessage), expectedBarMessage)
	}

	select {
	case barPublishResult := <-resultChannel:
		if barPublishResult != ":1\r\n" {
			t.Fatalf("HandlePublish(bar) = %q, expected %q", barPublishResult, ":1\r\n")
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for publish result")
	}

	fooClientConnection.SetReadDeadline(time.Now().Add(50 * time.Millisecond))
	_, readError = readAvailableBytes(fooClientConnection)
	if readError == nil {
		t.Fatal("foo client should not receive message for bar channel")
	}
}
