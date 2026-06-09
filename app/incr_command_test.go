package main

import "testing"

func resetIncrTestState(t *testing.T) {
	t.Helper()
	GetInstance().cache = make(map[string]CacheItem)
}

func TestParseIncrCommandArguments(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectedKey   string
		expectedError string
	}{
		{
			name:        "parses key argument",
			args:        []string{"foo"},
			expectedKey: "foo",
		},
		{
			name:          "rejects no arguments",
			args:          []string{},
			expectedError: "-ERR wrong number of arguments for 'incr' command\r\n",
		},
		{
			name:          "rejects too many arguments",
			args:          []string{"foo", "bar"},
			expectedError: "-ERR wrong number of arguments for 'incr' command\r\n",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			key, errorResponse := parseIncrCommandArguments(&RedisCommand{
				Type: CmdINCR,
				Args: testCase.args,
			})

			if errorResponse != testCase.expectedError {
				t.Errorf("parseIncrCommandArguments() error = %q, expected %q", errorResponse, testCase.expectedError)
			}

			if key != testCase.expectedKey {
				t.Errorf("parseIncrCommandArguments() key = %q, expected %q", key, testCase.expectedKey)
			}
		})
	}
}

func TestHandleIncrIncrementsExistingNumericValue(t *testing.T) {
	resetIncrTestState(t)

	HandleSet(&RedisCommand{
		Type: CmdSET,
		Args: []string{"foo", "41"},
	})

	result := HandleIncr(&RedisCommand{
		Type: CmdINCR,
		Args: []string{"foo"},
	})

	if result != ":42\r\n" {
		t.Errorf("HandleIncr() = %q, expected %q", result, ":42\r\n")
	}
}

func TestHandleIncrIncrementsValueOnRepeatedCalls(t *testing.T) {
	resetIncrTestState(t)

	HandleSet(&RedisCommand{
		Type: CmdSET,
		Args: []string{"foo", "5"},
	})

	firstResult := HandleIncr(&RedisCommand{
		Type: CmdINCR,
		Args: []string{"foo"},
	})
	if firstResult != ":6\r\n" {
		t.Errorf("first HandleIncr() = %q, expected %q", firstResult, ":6\r\n")
	}

	secondResult := HandleIncr(&RedisCommand{
		Type: CmdINCR,
		Args: []string{"foo"},
	})
	if secondResult != ":7\r\n" {
		t.Errorf("second HandleIncr() = %q, expected %q", secondResult, ":7\r\n")
	}
}

func TestHandleIncrSetsMissingKeyToOne(t *testing.T) {
	resetIncrTestState(t)

	tests := []struct {
		name string
		key  string
	}{
		{
			name: "codecrafters stage sets foo to 1 when key is missing",
			key:  "foo",
		},
		{
			name: "codecrafters stage sets bar to 1 when key is missing",
			key:  "bar",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := HandleIncr(&RedisCommand{
				Type: CmdINCR,
				Args: []string{testCase.key},
			})

			if result != ":1\r\n" {
				t.Errorf("HandleIncr() = %q, expected %q", result, ":1\r\n")
			}

			getResult := HandleGet(&RedisCommand{
				Type: CmdGET,
				Args: []string{testCase.key},
			})
			if getResult != "$1\r\n1\r\n" {
				t.Errorf("HandleGet() = %q, expected %q", getResult, "$1\r\n1\r\n")
			}
		})
	}
}

func TestHandleIncrArgumentParsing(t *testing.T) {
	resetIncrTestState(t)

	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "invalid argument count returns error",
			args:     []string{},
			expected: "-ERR wrong number of arguments for 'incr' command\r\n",
		},
		{
			name:     "too many arguments returns error",
			args:     []string{"foo", "bar"},
			expected: "-ERR wrong number of arguments for 'incr' command\r\n",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := HandleIncr(&RedisCommand{
				Type: CmdINCR,
				Args: testCase.args,
			})

			if result != testCase.expected {
				t.Errorf("HandleIncr() = %q, expected %q", result, testCase.expected)
			}
		})
	}
}
