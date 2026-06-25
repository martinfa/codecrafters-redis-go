package main

import (
	"fmt"
	"strings"
	"testing"
)

func formatExpectedZrangeResponse(members ...string) string {
	if len(members) == 0 {
		return "*0\r\n"
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("*%d\r\n", len(members)))
	for _, member := range members {
		builder.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(member), member))
	}

	return builder.String()
}

func seedCodecraftersZsetExample(t *testing.T) {
	t.Helper()
	resetSortedSetTestState(t)

	addSortedSetMembers(t, "zset_key",
		SortedSetMember{Member: "foo", Score: 100.0},
		SortedSetMember{Member: "bar", Score: 100.0},
		SortedSetMember{Member: "baz", Score: 20.0},
		SortedSetMember{Member: "caz", Score: 30.1},
		SortedSetMember{Member: "paz", Score: 40.2},
	)
}

func TestParseCommandType_ZRANGEReturnsKnownCommand(t *testing.T) {
	commandType := ParseCommandType("ZRANGE")
	if commandType != CmdZRANGE {
		t.Fatalf("expected ZRANGE to parse as CmdZRANGE, got %v", commandType)
	}
}

func TestParseZrangeCommandArguments(t *testing.T) {
	tests := []struct {
		name               string
		args               []string
		expectedKey        string
		expectedStartIndex string
		expectedStopIndex  string
		expectedError      string
	}{
		{
			name:               "parses key and indexes",
			args:               []string{"zset_key", "0", "2"},
			expectedKey:        "zset_key",
			expectedStartIndex: "0",
			expectedStopIndex:  "2",
		},
		{
			name:          "rejects missing arguments",
			args:          []string{},
			expectedError: "-ERR wrong number of arguments for 'zrange' command\r\n",
		},
		{
			name:          "rejects too few arguments",
			args:          []string{"zset_key", "0"},
			expectedError: "-ERR wrong number of arguments for 'zrange' command\r\n",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			key, startIndex, stopIndex, errorResponse := parseZrangeCommandArguments(&RedisCommand{
				Type: CmdZRANGE,
				Args: testCase.args,
			})

			if errorResponse != testCase.expectedError {
				t.Errorf("parseZrangeCommandArguments() error = %q, expected %q", errorResponse, testCase.expectedError)
			}

			if key != testCase.expectedKey {
				t.Errorf("parseZrangeCommandArguments() key = %q, expected %q", key, testCase.expectedKey)
			}

			if startIndex != testCase.expectedStartIndex {
				t.Errorf("parseZrangeCommandArguments() startIndex = %q, expected %q", startIndex, testCase.expectedStartIndex)
			}

			if stopIndex != testCase.expectedStopIndex {
				t.Errorf("parseZrangeCommandArguments() stopIndex = %q, expected %q", stopIndex, testCase.expectedStopIndex)
			}
		})
	}
}

func TestHandleZrangeCodecraftersStageExample(t *testing.T) {
	seedCodecraftersZsetExample(t)

	result := HandleZrange(&RedisCommand{
		Type: CmdZRANGE,
		Args: []string{"zset_key", "2", "4"},
	})

	expected := formatExpectedZrangeResponse("paz", "bar", "foo")
	if result != expected {
		t.Errorf("HandleZrange() = %q, expected %q", result, expected)
	}
}

func TestHandleZrangeStageDocumentationExample(t *testing.T) {
	resetSortedSetTestState(t)

	addSortedSetMembers(t, "racer_scores",
		SortedSetMember{Member: "Sam-Bodden", Score: 8.1},
		SortedSetMember{Member: "Royce", Score: 10.2},
		SortedSetMember{Member: "Ford", Score: 6.0},
		SortedSetMember{Member: "Prickett", Score: 14.1},
	)

	result := HandleZrange(&RedisCommand{
		Type: CmdZRANGE,
		Args: []string{"racer_scores", "0", "2"},
	})

	expected := formatExpectedZrangeResponse("Ford", "Sam-Bodden", "Royce")
	if result != expected {
		t.Errorf("HandleZrange() = %q, expected %q", result, expected)
	}
}

func TestHandleZrangeMissingKeyReturnsEmptyArray(t *testing.T) {
	resetSortedSetTestState(t)

	result := HandleZrange(&RedisCommand{
		Type: CmdZRANGE,
		Args: []string{"missing_key", "0", "2"},
	})

	if result != "*0\r\n" {
		t.Errorf("HandleZrange() = %q, expected %q", result, "*0\r\n")
	}
}

func TestHandleZrangeStartIndexGreaterThanStopIndexReturnsEmptyArray(t *testing.T) {
	seedCodecraftersZsetExample(t)

	result := HandleZrange(&RedisCommand{
		Type: CmdZRANGE,
		Args: []string{"zset_key", "4", "2"},
	})

	if result != "*0\r\n" {
		t.Errorf("HandleZrange() = %q, expected %q", result, "*0\r\n")
	}
}

func TestHandleZrangeStartIndexGreaterThanCardinalityReturnsEmptyArray(t *testing.T) {
	seedCodecraftersZsetExample(t)

	result := HandleZrange(&RedisCommand{
		Type: CmdZRANGE,
		Args: []string{"zset_key", "5", "6"},
	})

	if result != "*0\r\n" {
		t.Errorf("HandleZrange() = %q, expected %q", result, "*0\r\n")
	}
}

func TestHandleZrangeStopIndexGreaterThanCardinalityClampsToLastElement(t *testing.T) {
	seedCodecraftersZsetExample(t)

	result := HandleZrange(&RedisCommand{
		Type: CmdZRANGE,
		Args: []string{"zset_key", "3", "99"},
	})

	expected := formatExpectedZrangeResponse("bar", "foo")
	if result != expected {
		t.Errorf("HandleZrange() = %q, expected %q", result, expected)
	}
}

func seedRacerScoresExample(t *testing.T) {
	t.Helper()
	resetSortedSetTestState(t)

	addSortedSetMembers(t, "racer_scores",
		SortedSetMember{Member: "Sam-Bodden", Score: 8.5},
		SortedSetMember{Member: "Royce", Score: 10.2},
		SortedSetMember{Member: "Ford", Score: 6.1},
		SortedSetMember{Member: "Prickett", Score: 14.9},
		SortedSetMember{Member: "Ben", Score: 10.2},
	)
}

func TestHandleZrangeCodecraftersNegativeIndexStageExample(t *testing.T) {
	seedCodecraftersZsetExample(t)

	result := HandleZrange(&RedisCommand{
		Type: CmdZRANGE,
		Args: []string{"zset_key", "2", "-1"},
	})

	expected := formatExpectedZrangeResponse("paz", "bar", "foo")
	if result != expected {
		t.Errorf("HandleZrange() = %q, expected %q", result, expected)
	}
}

func TestHandleZrangeNegativeIndexes(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T)
		args     []string
		expected string
	}{
		{
			name:     "returns last two elements for indexes -2 and -1",
			setup:    seedRacerScoresExample,
			args:     []string{"racer_scores", "-2", "-1"},
			expected: formatExpectedZrangeResponse("Royce", "Prickett"),
		},
		{
			name:     "returns all items except last two for indexes 0 and -3",
			setup:    seedRacerScoresExample,
			args:     []string{"racer_scores", "0", "-3"},
			expected: formatExpectedZrangeResponse("Ford", "Sam-Bodden", "Ben"),
		},
		{
			name:     "returns single last element for indexes -1 and -1",
			setup:    seedRacerScoresExample,
			args:     []string{"racer_scores", "-1", "-1"},
			expected: formatExpectedZrangeResponse("Prickett"),
		},
		{
			name:     "treats out of range negative start index as zero",
			setup:    seedRacerScoresExample,
			args:     []string{"racer_scores", "-6", "2"},
			expected: formatExpectedZrangeResponse("Ford", "Sam-Bodden", "Ben"),
		},
		{
			name:     "returns empty array when normalized start is greater than stop",
			setup:    seedRacerScoresExample,
			args:     []string{"racer_scores", "-1", "-3"},
			expected: "*0\r\n",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(t)

			result := HandleZrange(&RedisCommand{
				Type: CmdZRANGE,
				Args: testCase.args,
			})

			if result != testCase.expected {
				t.Errorf("HandleZrange() = %q, expected %q", result, testCase.expected)
			}
		})
	}
}

func TestHandleZrangeArgumentParsing(t *testing.T) {
	resetSortedSetTestState(t)

	result := HandleZrange(&RedisCommand{
		Type: CmdZRANGE,
		Args: []string{"zset_key", "0"},
	})

	if result != "-ERR wrong number of arguments for 'zrange' command\r\n" {
		t.Errorf("HandleZrange() = %q, expected %q", result, "-ERR wrong number of arguments for 'zrange' command\r\n")
	}
}
