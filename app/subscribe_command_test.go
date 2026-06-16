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

func TestHandleConnectionCommandRejectsDisallowedCommandsInSubscribedMode(t *testing.T) {
	resetSubscribeTestState(t)
	ResetConnectionTransactionStatesForTest()

	connection := testConnection(t)

	HandleSubscribe(connection, &RedisCommand{
		Type: CmdSUBSCRIBE,
		Args: []string{"foo"},
	})

	tests := []struct {
		name        string
		command     *RedisCommand
		commandName string
	}{
		{
			name:        "rejects echo",
			command:     &RedisCommand{Type: CmdECHO, Args: []string{"hey"}},
			commandName: "echo",
		},
		{
			name:        "rejects set",
			command:     &RedisCommand{Type: CmdSET, Args: []string{"key", "value"}},
			commandName: "set",
		},
		{
			name:        "rejects get",
			command:     &RedisCommand{Type: CmdGET, Args: []string{"key"}},
			commandName: "get",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := HandleConnectionCommand(connection, testCase.command)

			expectedPrefix := "-ERR Can't execute '" + testCase.commandName + "'"
			if len(result) < len(expectedPrefix) || result[:len(expectedPrefix)] != expectedPrefix {
				t.Errorf("HandleConnectionCommand() = %q, expected prefix %q", result, expectedPrefix)
			}
		})
	}
}

func TestHandleConnectionCommandAllowsSubscribeAndPingInSubscribedMode(t *testing.T) {
	resetSubscribeTestState(t)
	ResetConnectionTransactionStatesForTest()

	connection := testConnection(t)

	HandleSubscribe(connection, &RedisCommand{
		Type: CmdSUBSCRIBE,
		Args: []string{"foo"},
	})

	secondSubscribeResult := HandleConnectionCommand(connection, &RedisCommand{
		Type: CmdSUBSCRIBE,
		Args: []string{"bar"},
	})
	expectedSecondSubscribe := formatExpectedSubscribeResponse("bar", 2)
	if secondSubscribeResult != expectedSecondSubscribe {
		t.Errorf("second SUBSCRIBE = %q, expected %q", secondSubscribeResult, expectedSecondSubscribe)
	}

	pingResult := HandleConnectionCommand(connection, &RedisCommand{
		Type: CmdPING,
		Args: []string{},
	})
	if pingResult != subscribedModePingResponse {
		t.Errorf("PING in subscribed mode = %q, expected %q", pingResult, subscribedModePingResponse)
	}
}

func TestHandlePingReturnsNormalResponseWhenNotSubscribed(t *testing.T) {
	resetSubscribeTestState(t)

	connection := testConnection(t)

	pingResult := HandleConnectionCommand(connection, &RedisCommand{
		Type: CmdPING,
		Args: []string{},
	})
	if pingResult != normalPingResponse {
		t.Errorf("PING before subscribe = %q, expected %q", pingResult, normalPingResponse)
	}
}

func TestHandlePingReturnsSubscribedModeResponseAfterSubscribe(t *testing.T) {
	resetSubscribeTestState(t)

	connection := testConnection(t)

	HandleSubscribe(connection, &RedisCommand{
		Type: CmdSUBSCRIBE,
		Args: []string{"foo"},
	})

	pingResult := HandleConnectionCommand(connection, &RedisCommand{
		Type: CmdPING,
		Args: []string{},
	})
	if pingResult != subscribedModePingResponse {
		t.Errorf("PING in subscribed mode = %q, expected %q", pingResult, subscribedModePingResponse)
	}
}

func TestHandleConnectionCommandAllowsCommandsBeforeSubscribe(t *testing.T) {
	resetSubscribeTestState(t)
	ResetConnectionTransactionStatesForTest()

	connection := testConnection(t)

	echoResult := HandleConnectionCommand(connection, &RedisCommand{
		Type: CmdECHO,
		Args: []string{"hey"},
	})
	if echoResult != "$3\r\nhey\r\n" {
		t.Errorf("ECHO before subscribe = %q, expected %q", echoResult, "$3\r\nhey\r\n")
	}
}
