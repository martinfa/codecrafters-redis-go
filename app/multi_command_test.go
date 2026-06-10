package main

import (
	"net"
	"testing"
)

func testConnection(t *testing.T) net.Conn {
	t.Helper()

	serverConnection, clientConnection := net.Pipe()
	t.Cleanup(func() {
		serverConnection.Close()
		clientConnection.Close()
	})

	return serverConnection
}

func TestParseMultiCommandArguments(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectedError string
	}{
		{
			name: "accepts no arguments",
			args: []string{},
		},
		{
			name:          "rejects extra arguments",
			args:          []string{"extra"},
			expectedError: "-ERR wrong number of arguments for 'multi' command\r\n",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			errorResponse := parseMultiCommandArguments(&RedisCommand{
				Type: CmdMULTI,
				Args: testCase.args,
			})

			if errorResponse != testCase.expectedError {
				t.Errorf("parseMultiCommandArguments() error = %q, expected %q", errorResponse, testCase.expectedError)
			}
		})
	}
}

func TestHandleMultiReturnsOK(t *testing.T) {
	ResetConnectionTransactionStatesForTest()
	connection := testConnection(t)

	result := HandleMulti(connection, &RedisCommand{
		Type: CmdMULTI,
		Args: []string{},
	})

	if result != "+OK\r\n" {
		t.Errorf("HandleMulti() = %q, expected %q", result, "+OK\r\n")
	}
}

func TestHandleMultiArgumentParsing(t *testing.T) {
	ResetConnectionTransactionStatesForTest()
	connection := testConnection(t)

	result := HandleMulti(connection, &RedisCommand{
		Type: CmdMULTI,
		Args: []string{"extra"},
	})

	if result != "-ERR wrong number of arguments for 'multi' command\r\n" {
		t.Errorf("HandleMulti() = %q, expected %q", result, "-ERR wrong number of arguments for 'multi' command\r\n")
	}
}

func TestParseExecCommandArguments(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectedError string
	}{
		{
			name: "accepts no arguments",
			args: []string{},
		},
		{
			name:          "rejects extra arguments",
			args:          []string{"extra"},
			expectedError: "-ERR wrong number of arguments for 'exec' command\r\n",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			errorResponse := parseExecCommandArguments(&RedisCommand{
				Type: CmdEXEC,
				Args: testCase.args,
			})

			if errorResponse != testCase.expectedError {
				t.Errorf("parseExecCommandArguments() error = %q, expected %q", errorResponse, testCase.expectedError)
			}
		})
	}
}

func TestHandleExecWithoutMultiReturnsError(t *testing.T) {
	ResetConnectionTransactionStatesForTest()
	connection := testConnection(t)

	result := HandleExec(connection, &RedisCommand{
		Type: CmdEXEC,
		Args: []string{},
	})

	if result != errExecWithoutMulti {
		t.Errorf("HandleExec() = %q, expected %q", result, errExecWithoutMulti)
	}
}

func TestHandleExecArgumentParsing(t *testing.T) {
	ResetConnectionTransactionStatesForTest()
	connection := testConnection(t)

	result := HandleExec(connection, &RedisCommand{
		Type: CmdEXEC,
		Args: []string{"extra"},
	})

	if result != "-ERR wrong number of arguments for 'exec' command\r\n" {
		t.Errorf("HandleExec() = %q, expected %q", result, "-ERR wrong number of arguments for 'exec' command\r\n")
	}
}
