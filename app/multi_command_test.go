package main

import "testing"

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
	result := HandleMulti(&RedisCommand{
		Type: CmdMULTI,
		Args: []string{},
	})

	if result != "+OK\r\n" {
		t.Errorf("HandleMulti() = %q, expected %q", result, "+OK\r\n")
	}
}

func TestHandleMultiArgumentParsing(t *testing.T) {
	result := HandleMulti(&RedisCommand{
		Type: CmdMULTI,
		Args: []string{"extra"},
	})

	if result != "-ERR wrong number of arguments for 'multi' command\r\n" {
		t.Errorf("HandleMulti() = %q, expected %q", result, "-ERR wrong number of arguments for 'multi' command\r\n")
	}
}
