package main

import (
	"fmt"
	"testing"
)

func resetLpopTestState(t *testing.T) {
	t.Helper()
	GetInstance().cache = make(map[string]CacheItem)
}

func formatExpectedBulkStringResponse(value string) string {
	return fmt.Sprintf("$%d\r\n%s\r\n", len(value), value)
}

func TestParseLpopCommandArguments(t *testing.T) {
	tests := []struct {
		name            string
		args            []string
		expectedListKey string
		expectedError   string
	}{
		{
			name:            "parses list key argument",
			args:            []string{"list_key"},
			expectedListKey: "list_key",
		},
		{
			name:          "rejects no arguments",
			args:          []string{},
			expectedError: "-ERR wrong number of arguments for 'lpop' command\r\n",
		},
		{
			name:          "rejects too many arguments",
			args:          []string{"list_key", "extra"},
			expectedError: "-ERR wrong number of arguments for 'lpop' command\r\n",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			listKey, errorResponse := parseLpopCommandArguments(&RedisCommand{
				Type: CmdLPOP,
				Args: testCase.args,
			})

			if errorResponse != testCase.expectedError {
				t.Errorf("parseLpopCommandArguments() error = %q, expected %q", errorResponse, testCase.expectedError)
			}

			if listKey != testCase.expectedListKey {
				t.Errorf("parseLpopCommandArguments() listKey = %q, expected %q", listKey, testCase.expectedListKey)
			}
		})
	}
}

func TestHandleLpopReturnsFirstElement(t *testing.T) {
	resetLpopTestState(t)

	rpushResult := HandleRpush(&RedisCommand{
		Type: CmdRPUSH,
		Args: []string{"list_key", "one", "two", "three", "four", "five"},
	})
	if rpushResult != ":5\r\n" {
		t.Fatalf("HandleRpush() = %q, expected %q", rpushResult, ":5\r\n")
	}

	result := HandleLpop(&RedisCommand{
		Type: CmdLPOP,
		Args: []string{"list_key"},
	})

	expected := formatExpectedBulkStringResponse("one")
	if result != expected {
		t.Errorf("HandleLpop() = %q, expected %q", result, expected)
	}
}

func TestHandleLpopRemovesFirstElementFromList(t *testing.T) {
	resetLpopTestState(t)

	HandleRpush(&RedisCommand{
		Type: CmdRPUSH,
		Args: []string{"list_key", "one", "two", "three", "four", "five"},
	})

	popResult := HandleLpop(&RedisCommand{
		Type: CmdLPOP,
		Args: []string{"list_key"},
	})
	if popResult != formatExpectedBulkStringResponse("one") {
		t.Fatalf("HandleLpop() = %q, expected %q", popResult, formatExpectedBulkStringResponse("one"))
	}

	rangeResult := HandleLrange(&RedisCommand{
		Type: CmdLRANGE,
		Args: []string{"list_key", "0", "-1"},
	})
	expectedRangeResult := formatExpectedLrangeResponse("two", "three", "four", "five")
	if rangeResult != expectedRangeResult {
		t.Errorf("HandleLrange() = %q, expected %q", rangeResult, expectedRangeResult)
	}
}

func TestHandleLpopCodecraftersStageExample(t *testing.T) {
	resetLpopTestState(t)

	HandleRpush(&RedisCommand{
		Type: CmdRPUSH,
		Args: []string{"list_key", "a", "b", "c", "d"},
	})

	popResult := HandleLpop(&RedisCommand{
		Type: CmdLPOP,
		Args: []string{"list_key"},
	})
	if popResult != formatExpectedBulkStringResponse("a") {
		t.Fatalf("HandleLpop() = %q, expected %q", popResult, formatExpectedBulkStringResponse("a"))
	}
}

func TestHandleLpopReturnsNullBulkStringForMissingList(t *testing.T) {
	resetLpopTestState(t)

	result := HandleLpop(&RedisCommand{
		Type: CmdLPOP,
		Args: []string{"missing_list_key"},
	})

	if result != "$-1\r\n" {
		t.Errorf("HandleLpop() = %q, expected %q", result, "$-1\r\n")
	}
}

func TestHandleLpopReturnsNullBulkStringForEmptyList(t *testing.T) {
	resetLpopTestState(t)

	HandleRpush(&RedisCommand{
		Type: CmdRPUSH,
		Args: []string{"list_key", "only"},
	})

	firstPopResult := HandleLpop(&RedisCommand{
		Type: CmdLPOP,
		Args: []string{"list_key"},
	})
	if firstPopResult != formatExpectedBulkStringResponse("only") {
		t.Fatalf("first HandleLpop() = %q, expected %q", firstPopResult, formatExpectedBulkStringResponse("only"))
	}

	secondPopResult := HandleLpop(&RedisCommand{
		Type: CmdLPOP,
		Args: []string{"list_key"},
	})
	if secondPopResult != "$-1\r\n" {
		t.Errorf("second HandleLpop() = %q, expected %q", secondPopResult, "$-1\r\n")
	}
}

func TestHandleLpopArgumentParsing(t *testing.T) {
	resetLpopTestState(t)

	result := HandleLpop(&RedisCommand{
		Type: CmdLPOP,
		Args: []string{},
	})

	if result != "-ERR wrong number of arguments for 'lpop' command\r\n" {
		t.Errorf("HandleLpop() = %q, expected %q", result, "-ERR wrong number of arguments for 'lpop' command\r\n")
	}
}
