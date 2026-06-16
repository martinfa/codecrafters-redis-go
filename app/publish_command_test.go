package main

import "testing"

func resetPublishTestState(t *testing.T) {
	t.Helper()
	ResetConnectionPubSubStatesForTest()
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

	firstConnection := testConnection(t)
	secondConnection := testConnection(t)
	thirdConnection := testConnection(t)

	HandleSubscribe(firstConnection, &RedisCommand{
		Type: CmdSUBSCRIBE,
		Args: []string{"foo"},
	})
	HandleSubscribe(secondConnection, &RedisCommand{
		Type: CmdSUBSCRIBE,
		Args: []string{"bar"},
	})
	HandleSubscribe(thirdConnection, &RedisCommand{
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

	connection := testConnection(t)

	HandleSubscribe(connection, &RedisCommand{
		Type: CmdSUBSCRIBE,
		Args: []string{"foo"},
	})
	HandleSubscribe(connection, &RedisCommand{
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
