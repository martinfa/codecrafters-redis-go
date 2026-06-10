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

func TestHandleExecEmptyTransaction(t *testing.T) {
	ResetConnectionTransactionStatesForTest()
	connection := testConnection(t)

	multiResult := HandleMulti(connection, &RedisCommand{
		Type: CmdMULTI,
		Args: []string{},
	})
	if multiResult != "+OK\r\n" {
		t.Fatalf("HandleMulti() = %q, expected %q", multiResult, "+OK\r\n")
	}

	firstExecResult := HandleExec(connection, &RedisCommand{
		Type: CmdEXEC,
		Args: []string{},
	})
	if firstExecResult != "*0\r\n" {
		t.Errorf("first HandleExec() = %q, expected %q", firstExecResult, "*0\r\n")
	}

	secondExecResult := HandleExec(connection, &RedisCommand{
		Type: CmdEXEC,
		Args: []string{},
	})
	if secondExecResult != errExecWithoutMulti {
		t.Errorf("second HandleExec() = %q, expected %q", secondExecResult, errExecWithoutMulti)
	}
}

func resetTransactionTestState(t *testing.T) {
	t.Helper()
	GetInstance().cache = make(map[string]CacheItem)
	ResetConnectionTransactionStatesForTest()
}

func TestTransactionQueuesSetAndIncrWithoutExecutingThem(t *testing.T) {
	resetTransactionTestState(t)

	transactionConnection := testConnection(t)
	_ = testConnection(t)

	multiResult := HandleConnectionCommand(transactionConnection, &RedisCommand{
		Type: CmdMULTI,
		Args: []string{},
	})
	if multiResult != "+OK\r\n" {
		t.Fatalf("HandleConnectionCommand(MULTI) = %q, expected %q", multiResult, "+OK\r\n")
	}

	setResult := HandleConnectionCommand(transactionConnection, &RedisCommand{
		Type: CmdSET,
		Args: []string{"foo", "41"},
	})
	if setResult != queuedCommandResponse {
		t.Errorf("HandleConnectionCommand(SET) = %q, expected %q", setResult, queuedCommandResponse)
	}

	incrResult := HandleConnectionCommand(transactionConnection, &RedisCommand{
		Type: CmdINCR,
		Args: []string{"foo"},
	})
	if incrResult != queuedCommandResponse {
		t.Errorf("HandleConnectionCommand(INCR) = %q, expected %q", incrResult, queuedCommandResponse)
	}

	getResult := HandleGet(&RedisCommand{
		Type: CmdGET,
		Args: []string{"foo"},
	})
	if getResult != "$-1\r\n" {
		t.Errorf("HandleGet(foo) = %q, expected %q", getResult, "$-1\r\n")
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
