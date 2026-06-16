package main

import (
	"testing"
	"time"
)

func resetUnsubscribeTestState(t *testing.T) {
	t.Helper()
	ResetConnectionPubSubStatesForTest()
	ResetConnectionWriteMutexesForTest()
}

func formatExpectedUnsubscribeResponse(channel string, remainingSubscriptionCount int) string {
	return encodeUnsubscribeResponse(channel, remainingSubscriptionCount)
}

func TestParseUnsubscribeCommandArguments(t *testing.T) {
	tests := []struct {
		name            string
		args            []string
		expectedChannel string
		expectedError   string
	}{
		{
			name:            "parses channel argument",
			args:            []string{"foo"},
			expectedChannel: "foo",
		},
		{
			name:          "rejects no arguments",
			args:          []string{},
			expectedError: "-ERR wrong number of arguments for 'unsubscribe' command\r\n",
		},
		{
			name:          "rejects too many arguments",
			args:          []string{"foo", "bar"},
			expectedError: "-ERR wrong number of arguments for 'unsubscribe' command\r\n",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			channel, errorResponse := parseUnsubscribeCommandArguments(&RedisCommand{
				Type: CmdUNSUBSCRIBE,
				Args: testCase.args,
			})

			if errorResponse != testCase.expectedError {
				t.Errorf("parseUnsubscribeCommandArguments() error = %q, expected %q", errorResponse, testCase.expectedError)
			}

			if channel != testCase.expectedChannel {
				t.Errorf("parseUnsubscribeCommandArguments() channel = %q, expected %q", channel, testCase.expectedChannel)
			}
		})
	}
}

func TestHandleUnsubscribeCodecraftersStageExample(t *testing.T) {
	resetUnsubscribeTestState(t)

	connection := testConnection(t)

	HandleSubscribe(connection, &RedisCommand{
		Type: CmdSUBSCRIBE,
		Args: []string{"foo"},
	})
	HandleSubscribe(connection, &RedisCommand{
		Type: CmdSUBSCRIBE,
		Args: []string{"bar"},
	})

	unsubscribeFooResult := HandleUnsubscribe(connection, &RedisCommand{
		Type: CmdUNSUBSCRIBE,
		Args: []string{"foo"},
	})
	if unsubscribeFooResult != formatExpectedUnsubscribeResponse("foo", 1) {
		t.Errorf("HandleUnsubscribe(foo) = %q, expected %q", unsubscribeFooResult, formatExpectedUnsubscribeResponse("foo", 1))
	}

	unsubscribeBarResult := HandleUnsubscribe(connection, &RedisCommand{
		Type: CmdUNSUBSCRIBE,
		Args: []string{"bar"},
	})
	if unsubscribeBarResult != formatExpectedUnsubscribeResponse("bar", 0) {
		t.Errorf("HandleUnsubscribe(bar) = %q, expected %q", unsubscribeBarResult, formatExpectedUnsubscribeResponse("bar", 0))
	}
}

func TestHandleUnsubscribeFromUnsubscribedChannelDoesNotChangeCount(t *testing.T) {
	resetUnsubscribeTestState(t)

	connection := testConnection(t)

	HandleSubscribe(connection, &RedisCommand{
		Type: CmdSUBSCRIBE,
		Args: []string{"foo"},
	})

	result := HandleUnsubscribe(connection, &RedisCommand{
		Type: CmdUNSUBSCRIBE,
		Args: []string{"bar"},
	})
	if result != formatExpectedUnsubscribeResponse("bar", 1) {
		t.Errorf("HandleUnsubscribe(bar) = %q, expected %q", result, formatExpectedUnsubscribeResponse("bar", 1))
	}
}

func TestHandleUnsubscribeLeavesSubscribedModeWhenLastChannelRemoved(t *testing.T) {
	resetUnsubscribeTestState(t)
	ResetConnectionTransactionStatesForTest()

	connection := testConnection(t)

	HandleSubscribe(connection, &RedisCommand{
		Type: CmdSUBSCRIBE,
		Args: []string{"foo"},
	})

	HandleUnsubscribe(connection, &RedisCommand{
		Type: CmdUNSUBSCRIBE,
		Args: []string{"foo"},
	})

	echoResult := HandleConnectionCommand(connection, &RedisCommand{
		Type: CmdECHO,
		Args: []string{"hey"},
	})
	if echoResult != "$3\r\nhey\r\n" {
		t.Errorf("ECHO after last unsubscribe = %q, expected %q", echoResult, "$3\r\nhey\r\n")
	}
}

func TestHandleUnsubscribeStopsMessageDeliveryForChannel(t *testing.T) {
	resetUnsubscribeTestState(t)

	serverConnection, clientConnection := testConnectionPair(t)

	HandleSubscribe(serverConnection, &RedisCommand{
		Type: CmdSUBSCRIBE,
		Args: []string{"foo"},
	})

	HandleUnsubscribe(serverConnection, &RedisCommand{
		Type: CmdUNSUBSCRIBE,
		Args: []string{"foo"},
	})

	resultChannel := make(chan string, 1)
	go func() {
		resultChannel <- HandlePublish(&RedisCommand{
			Type: CmdPUBLISH,
			Args: []string{"foo", "after-unsubscribe"},
		})
	}()

	select {
	case publishResult := <-resultChannel:
		if publishResult != ":0\r\n" {
			t.Fatalf("HandlePublish(foo) = %q, expected %q", publishResult, ":0\r\n")
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for publish result")
	}

	clientConnection.SetReadDeadline(time.Now().Add(50 * time.Millisecond))
	_, readError := readAvailableBytes(clientConnection)
	if readError == nil {
		t.Fatal("client should not receive message after unsubscribing from channel")
	}
}

func TestHandleUnsubscribeArgumentParsing(t *testing.T) {
	resetUnsubscribeTestState(t)

	connection := testConnection(t)

	result := HandleUnsubscribe(connection, &RedisCommand{
		Type: CmdUNSUBSCRIBE,
		Args: []string{},
	})

	if result != "-ERR wrong number of arguments for 'unsubscribe' command\r\n" {
		t.Errorf("HandleUnsubscribe() = %q, expected %q", result, "-ERR wrong number of arguments for 'unsubscribe' command\r\n")
	}
}
