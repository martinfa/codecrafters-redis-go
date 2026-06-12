package main

import (
	"fmt"
	"strings"
	"testing"
)

func resetLrangeTestState(t *testing.T) {
	t.Helper()
	GetInstance().cache = make(map[string]CacheItem)
}

func seedListKeyWithFiveElements(t *testing.T) {
	t.Helper()

	result := HandleRpush(&RedisCommand{
		Type: CmdRPUSH,
		Args: []string{"list_key", "a", "b", "c", "d", "e"},
	})
	if result != ":5\r\n" {
		t.Fatalf("HandleRpush() = %q, expected %q", result, ":5\r\n")
	}
}

func formatExpectedLrangeResponse(elements ...string) string {
	if len(elements) == 0 {
		return "*0\r\n"
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("*%d\r\n", len(elements)))
	for _, element := range elements {
		builder.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(element), element))
	}

	return builder.String()
}

func TestParseLrangeCommandArguments(t *testing.T) {
	tests := []struct {
		name              string
		args              []string
		expectedListKey   string
		expectedStartIndex string
		expectedStopIndex  string
		expectedError     string
	}{
		{
			name:               "parses list key and indexes",
			args:               []string{"list_key", "0", "2"},
			expectedListKey:    "list_key",
			expectedStartIndex: "0",
			expectedStopIndex:  "2",
		},
		{
			name:          "rejects no arguments",
			args:          []string{},
			expectedError: "-ERR wrong number of arguments for 'lrange' command\r\n",
		},
		{
			name:          "rejects missing stop index",
			args:          []string{"list_key", "0"},
			expectedError: "-ERR wrong number of arguments for 'lrange' command\r\n",
		},
		{
			name:          "rejects too many arguments",
			args:          []string{"list_key", "0", "2", "extra"},
			expectedError: "-ERR wrong number of arguments for 'lrange' command\r\n",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			listKey, startIndex, stopIndex, errorResponse := parseLrangeCommandArguments(&RedisCommand{
				Type: CmdLRANGE,
				Args: testCase.args,
			})

			if errorResponse != testCase.expectedError {
				t.Errorf("parseLrangeCommandArguments() error = %q, expected %q", errorResponse, testCase.expectedError)
			}

			if listKey != testCase.expectedListKey {
				t.Errorf("parseLrangeCommandArguments() listKey = %q, expected %q", listKey, testCase.expectedListKey)
			}

			if startIndex != testCase.expectedStartIndex {
				t.Errorf("parseLrangeCommandArguments() startIndex = %q, expected %q", startIndex, testCase.expectedStartIndex)
			}

			if stopIndex != testCase.expectedStopIndex {
				t.Errorf("parseLrangeCommandArguments() stopIndex = %q, expected %q", stopIndex, testCase.expectedStopIndex)
			}
		})
	}
}

func TestHandleLrange(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T)
		args     []string
		expected string
	}{
		{
			name:  "codecrafters stage returns first three elements",
			setup: seedListKeyWithFiveElements,
			args:  []string{"list_key", "0", "2"},
			expected: formatExpectedLrangeResponse("a", "b", "c"),
		},
		{
			name:  "returns first two elements for indexes 0 and 1",
			setup: seedListKeyWithFiveElements,
			args:  []string{"list_key", "0", "1"},
			expected: formatExpectedLrangeResponse("a", "b"),
		},
		{
			name:  "returns elements from index 2 through 4 inclusive",
			setup: seedListKeyWithFiveElements,
			args:  []string{"list_key", "2", "4"},
			expected: formatExpectedLrangeResponse("c", "d", "e"),
		},
		{
			name:  "returns empty array when list does not exist",
			setup: func(t *testing.T) {},
			args:  []string{"missing_list", "0", "2"},
			expected: "*0\r\n",
		},
		{
			name:  "returns empty array when start index equals list length",
			setup: seedListKeyWithFiveElements,
			args:  []string{"list_key", "5", "5"},
			expected: "*0\r\n",
		},
		{
			name:  "returns empty array when start index is greater than list length",
			setup: seedListKeyWithFiveElements,
			args:  []string{"list_key", "10", "10"},
			expected: "*0\r\n",
		},
		{
			name:  "treats stop index equal to list length as last element",
			setup: seedListKeyWithFiveElements,
			args:  []string{"list_key", "0", "5"},
			expected: formatExpectedLrangeResponse("a", "b", "c", "d", "e"),
		},
		{
			name:  "treats stop index beyond list end as last element",
			setup: seedListKeyWithFiveElements,
			args:  []string{"list_key", "2", "100"},
			expected: formatExpectedLrangeResponse("c", "d", "e"),
		},
		{
			name:  "returns empty array when start index is greater than stop index",
			setup: seedListKeyWithFiveElements,
			args:  []string{"list_key", "3", "1"},
			expected: "*0\r\n",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			resetLrangeTestState(t)
			testCase.setup(t)

			result := HandleLrange(&RedisCommand{
				Type: CmdLRANGE,
				Args: testCase.args,
			})

			if result != testCase.expected {
				t.Errorf("HandleLrange() = %q, expected %q", result, testCase.expected)
			}
		})
	}
}

func TestHandleLrangeArgumentParsing(t *testing.T) {
	resetLrangeTestState(t)

	result := HandleLrange(&RedisCommand{
		Type: CmdLRANGE,
		Args: []string{"list_key", "0"},
	})

	if result != "-ERR wrong number of arguments for 'lrange' command\r\n" {
		t.Errorf("HandleLrange() = %q, expected %q", result, "-ERR wrong number of arguments for 'lrange' command\r\n")
	}
}
