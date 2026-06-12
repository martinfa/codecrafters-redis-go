package main

import "testing"

func resetLpushTestState(t *testing.T) {
	t.Helper()
	GetInstance().cache = make(map[string]CacheItem)
}

func TestParseLpushCommandArguments(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		expectedListKey  string
		expectedElements []string
		expectedError    string
	}{
		{
			name:             "parses list key and elements",
			args:             []string{"list_key", "a", "b", "c"},
			expectedListKey:  "list_key",
			expectedElements: []string{"a", "b", "c"},
		},
		{
			name:          "rejects no arguments",
			args:          []string{},
			expectedError: "-ERR wrong number of arguments for 'lpush' command\r\n",
		},
		{
			name:          "rejects missing elements",
			args:          []string{"list_key"},
			expectedError: "-ERR wrong number of arguments for 'lpush' command\r\n",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			listKey, elements, errorResponse := parseLpushCommandArguments(&RedisCommand{
				Type: CmdLPUSH,
				Args: testCase.args,
			})

			if errorResponse != testCase.expectedError {
				t.Errorf("parseLpushCommandArguments() error = %q, expected %q", errorResponse, testCase.expectedError)
			}

			if listKey != testCase.expectedListKey {
				t.Errorf("parseLpushCommandArguments() listKey = %q, expected %q", listKey, testCase.expectedListKey)
			}

			if len(elements) != len(testCase.expectedElements) {
				t.Fatalf("parseLpushCommandArguments() elements = %v, expected %v", elements, testCase.expectedElements)
			}

			for index, element := range elements {
				if element != testCase.expectedElements[index] {
					t.Errorf("parseLpushCommandArguments() elements[%d] = %q, expected %q", index, element, testCase.expectedElements[index])
				}
			}
		})
	}
}

func TestHandleLpushReturnsListLength(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "returns length 1 for single element",
			args:     []string{"list_key", "c"},
			expected: ":1\r\n",
		},
		{
			name:     "returns length 3 for three elements",
			args:     []string{"list_key", "a", "b", "c"},
			expected: ":3\r\n",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			resetLpushTestState(t)

			result := HandleLpush(&RedisCommand{
				Type: CmdLPUSH,
				Args: testCase.args,
			})

			if result != testCase.expected {
				t.Errorf("HandleLpush() = %q, expected %q", result, testCase.expected)
			}
		})
	}
}

func TestHandleLpushPrependsElementsInReverseOrder(t *testing.T) {
	resetLpushTestState(t)

	lengthResult := HandleLpush(&RedisCommand{
		Type: CmdLPUSH,
		Args: []string{"list_key", "a", "b", "c"},
	})
	if lengthResult != ":3\r\n" {
		t.Fatalf("HandleLpush() = %q, expected %q", lengthResult, ":3\r\n")
	}

	rangeResult := HandleLrange(&RedisCommand{
		Type: CmdLRANGE,
		Args: []string{"list_key", "0", "-1"},
	})
	expectedRangeResult := formatExpectedLrangeResponse("c", "b", "a")
	if rangeResult != expectedRangeResult {
		t.Errorf("HandleLrange() = %q, expected %q", rangeResult, expectedRangeResult)
	}
}

func TestHandleLpushCodecraftersStageExample(t *testing.T) {
	resetLpushTestState(t)

	firstResult := HandleLpush(&RedisCommand{
		Type: CmdLPUSH,
		Args: []string{"list_key", "c"},
	})
	if firstResult != ":1\r\n" {
		t.Fatalf("first HandleLpush() = %q, expected %q", firstResult, ":1\r\n")
	}

	secondResult := HandleLpush(&RedisCommand{
		Type: CmdLPUSH,
		Args: []string{"list_key", "b", "a"},
	})
	if secondResult != ":3\r\n" {
		t.Fatalf("second HandleLpush() = %q, expected %q", secondResult, ":3\r\n")
	}

	rangeResult := HandleLrange(&RedisCommand{
		Type: CmdLRANGE,
		Args: []string{"list_key", "0", "-1"},
	})
	expectedRangeResult := formatExpectedLrangeResponse("a", "b", "c")
	if rangeResult != expectedRangeResult {
		t.Errorf("HandleLrange() = %q, expected %q", rangeResult, expectedRangeResult)
	}
}

func TestHandleLpushPrependsToExistingList(t *testing.T) {
	resetLpushTestState(t)

	rpushResult := HandleRpush(&RedisCommand{
		Type: CmdRPUSH,
		Args: []string{"list_key", "x", "y"},
	})
	if rpushResult != ":2\r\n" {
		t.Fatalf("HandleRpush() = %q, expected %q", rpushResult, ":2\r\n")
	}

	lpushResult := HandleLpush(&RedisCommand{
		Type: CmdLPUSH,
		Args: []string{"list_key", "a", "b"},
	})
	if lpushResult != ":4\r\n" {
		t.Fatalf("HandleLpush() = %q, expected %q", lpushResult, ":4\r\n")
	}

	rangeResult := HandleLrange(&RedisCommand{
		Type: CmdLRANGE,
		Args: []string{"list_key", "0", "-1"},
	})
	expectedRangeResult := formatExpectedLrangeResponse("b", "a", "x", "y")
	if rangeResult != expectedRangeResult {
		t.Errorf("HandleLrange() = %q, expected %q", rangeResult, expectedRangeResult)
	}
}

func TestHandleLpushArgumentParsing(t *testing.T) {
	resetLpushTestState(t)

	result := HandleLpush(&RedisCommand{
		Type: CmdLPUSH,
		Args: []string{"list_key"},
	})

	if result != "-ERR wrong number of arguments for 'lpush' command\r\n" {
		t.Errorf("HandleLpush() = %q, expected %q", result, "-ERR wrong number of arguments for 'lpush' command\r\n")
	}
}
