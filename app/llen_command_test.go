package main

import "testing"

func resetLlenTestState(t *testing.T) {
	t.Helper()
	GetInstance().cache = make(map[string]CacheItem)
}

func TestParseLlenCommandArguments(t *testing.T) {
	tests := []struct {
		name              string
		args              []string
		expectedListKey   string
		expectedError     string
	}{
		{
			name:            "parses list key argument",
			args:            []string{"list_key"},
			expectedListKey: "list_key",
		},
		{
			name:          "rejects no arguments",
			args:          []string{},
			expectedError: "-ERR wrong number of arguments for 'llen' command\r\n",
		},
		{
			name:          "rejects too many arguments",
			args:          []string{"list_key", "extra"},
			expectedError: "-ERR wrong number of arguments for 'llen' command\r\n",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			listKey, errorResponse := parseLlenCommandArguments(&RedisCommand{
				Type: CmdLLEN,
				Args: testCase.args,
			})

			if errorResponse != testCase.expectedError {
				t.Errorf("parseLlenCommandArguments() error = %q, expected %q", errorResponse, testCase.expectedError)
			}

			if listKey != testCase.expectedListKey {
				t.Errorf("parseLlenCommandArguments() listKey = %q, expected %q", listKey, testCase.expectedListKey)
			}
		})
	}
}

func TestHandleLlenReturnsListLengthAfterRpush(t *testing.T) {
	resetLlenTestState(t)

	rpushResult := HandleRpush(&RedisCommand{
		Type: CmdRPUSH,
		Args: []string{"list_key", "a", "b", "c", "d"},
	})
	if rpushResult != ":4\r\n" {
		t.Fatalf("HandleRpush() = %q, expected %q", rpushResult, ":4\r\n")
	}

	result := HandleLlen(&RedisCommand{
		Type: CmdLLEN,
		Args: []string{"list_key"},
	})

	if result != ":4\r\n" {
		t.Errorf("HandleLlen() = %q, expected %q", result, ":4\r\n")
	}
}

func TestHandleLlenReturnsListLengthAfterLpush(t *testing.T) {
	resetLlenTestState(t)

	lpushResult := HandleLpush(&RedisCommand{
		Type: CmdLPUSH,
		Args: []string{"list_key", "a", "b", "c"},
	})
	if lpushResult != ":3\r\n" {
		t.Fatalf("HandleLpush() = %q, expected %q", lpushResult, ":3\r\n")
	}

	result := HandleLlen(&RedisCommand{
		Type: CmdLLEN,
		Args: []string{"list_key"},
	})

	if result != ":3\r\n" {
		t.Errorf("HandleLlen() = %q, expected %q", result, ":3\r\n")
	}
}

func TestHandleLlenReturnsZeroForMissingList(t *testing.T) {
	resetLlenTestState(t)

	result := HandleLlen(&RedisCommand{
		Type: CmdLLEN,
		Args: []string{"missing_list_key"},
	})

	if result != ":0\r\n" {
		t.Errorf("HandleLlen() = %q, expected %q", result, ":0\r\n")
	}
}

func TestHandleLlenArgumentParsing(t *testing.T) {
	resetLlenTestState(t)

	result := HandleLlen(&RedisCommand{
		Type: CmdLLEN,
		Args: []string{},
	})

	if result != "-ERR wrong number of arguments for 'llen' command\r\n" {
		t.Errorf("HandleLlen() = %q, expected %q", result, "-ERR wrong number of arguments for 'llen' command\r\n")
	}
}
