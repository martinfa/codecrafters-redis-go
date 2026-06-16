package main

import "testing"

func resetSubscribeTestState(t *testing.T) {
	t.Helper()
	ResetConnectionPubSubStatesForTest()
}

func formatExpectedSubscribeResponse(channel string, subscriptionCount int) string {
	return encodeSubscribeResponse(channel, subscriptionCount)
}

func TestParseSubscribeCommandArguments(t *testing.T) {
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
			expectedError: "-ERR wrong number of arguments for 'subscribe' command\r\n",
		},
		{
			name:          "rejects too many arguments",
			args:          []string{"foo", "bar"},
			expectedError: "-ERR wrong number of arguments for 'subscribe' command\r\n",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			channel, errorResponse := parseSubscribeCommandArguments(&RedisCommand{
				Type: CmdSUBSCRIBE,
				Args: testCase.args,
			})

			if errorResponse != testCase.expectedError {
				t.Errorf("parseSubscribeCommandArguments() error = %q, expected %q", errorResponse, testCase.expectedError)
			}

			if channel != testCase.expectedChannel {
				t.Errorf("parseSubscribeCommandArguments() channel = %q, expected %q", channel, testCase.expectedChannel)
			}
		})
	}
}

func TestHandleSubscribeReturnsSubscribeConfirmation(t *testing.T) {
	resetSubscribeTestState(t)

	connection := testConnection(t)

	result := HandleSubscribe(connection, &RedisCommand{
		Type: CmdSUBSCRIBE,
		Args: []string{"foo"},
	})

	expected := formatExpectedSubscribeResponse("foo", 1)
	if result != expected {
		t.Errorf("HandleSubscribe() = %q, expected %q", result, expected)
	}
}

func TestHandleSubscribeCodecraftersStageExample(t *testing.T) {
	resetSubscribeTestState(t)

	connection := testConnection(t)

	result := HandleSubscribe(connection, &RedisCommand{
		Type: CmdSUBSCRIBE,
		Args: []string{"mychan"},
	})

	expected := "*3\r\n$9\r\nsubscribe\r\n$6\r\nmychan\r\n:1\r\n"
	if result != expected {
		t.Errorf("HandleSubscribe() = %q, expected %q", result, expected)
	}
}

func TestHandleSubscribeArgumentParsing(t *testing.T) {
	resetSubscribeTestState(t)

	connection := testConnection(t)

	result := HandleSubscribe(connection, &RedisCommand{
		Type: CmdSUBSCRIBE,
		Args: []string{},
	})

	if result != "-ERR wrong number of arguments for 'subscribe' command\r\n" {
		t.Errorf("HandleSubscribe() = %q, expected %q", result, "-ERR wrong number of arguments for 'subscribe' command\r\n")
	}
}
