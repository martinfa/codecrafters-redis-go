package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestParseXrangeCommandArguments(t *testing.T) {
	tests := []struct {
		name              string
		args              []string
		expectedStreamKey string
		expectedStartID   string
		expectedEndID     string
		expectedError     string
	}{
		{
			name:              "parses stream key and full entry ids",
			args:              []string{"stream_key", "0-2", "0-3"},
			expectedStreamKey: "stream_key",
			expectedStartID:   "0-2",
			expectedEndID:     "0-3",
		},
		{
			name:              "parses ids without sequence numbers",
			args:              []string{"some_key", "1526985054069", "1526985054079"},
			expectedStreamKey: "some_key",
			expectedStartID:   "1526985054069",
			expectedEndID:     "1526985054079",
		},
		{
			name:              "parses mixed id formats",
			args:              []string{"stream_key", "1526985054069", "1526985054079-0"},
			expectedStreamKey: "stream_key",
			expectedStartID:   "1526985054069",
			expectedEndID:     "1526985054079-0",
		},
		{
			name:          "rejects no arguments",
			args:          []string{},
			expectedError: "-ERR wrong number of arguments for 'xrange' command\r\n",
		},
		{
			name:          "rejects stream key only",
			args:          []string{"stream_key"},
			expectedError: "-ERR wrong number of arguments for 'xrange' command\r\n",
		},
		{
			name:          "rejects missing end id",
			args:          []string{"stream_key", "0-2"},
			expectedError: "-ERR wrong number of arguments for 'xrange' command\r\n",
		},
		{
			name:          "rejects too many arguments",
			args:          []string{"stream_key", "0-2", "0-3", "extra"},
			expectedError: "-ERR wrong number of arguments for 'xrange' command\r\n",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			command := &RedisCommand{
				Type: CmdXRANGE,
				Args: testCase.args,
			}

			streamKey, startID, endID, errorResponse := parseXrangeCommandArguments(command)

			if errorResponse != testCase.expectedError {
				t.Errorf("parseXrangeCommandArguments() error = %q, expected %q", errorResponse, testCase.expectedError)
			}

			if streamKey != testCase.expectedStreamKey {
				t.Errorf("parseXrangeCommandArguments() streamKey = %q, expected %q", streamKey, testCase.expectedStreamKey)
			}

			if startID != testCase.expectedStartID {
				t.Errorf("parseXrangeCommandArguments() startID = %q, expected %q", startID, testCase.expectedStartID)
			}

			if endID != testCase.expectedEndID {
				t.Errorf("parseXrangeCommandArguments() endID = %q, expected %q", endID, testCase.expectedEndID)
			}
		})
	}
}

func TestHandleXrangeArgumentParsing(t *testing.T) {
	tests := []struct {
		name     string
		cmd      *RedisCommand
		expected string
	}{
		{
			name: "invalid argument count returns error",
			cmd: &RedisCommand{
				Type: CmdXRANGE,
				Args: []string{"stream_key", "0-2"},
			},
			expected: "-ERR wrong number of arguments for 'xrange' command\r\n",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := HandleXrange(testCase.cmd)
			if result != testCase.expected {
				t.Errorf("HandleXrange() = %q, expected %q", result, testCase.expected)
			}
		})
	}
}

func formatRESPBulkString(value string) string {
	return fmt.Sprintf("$%d\r\n%s\r\n", len(value), value)
}

func formatExpectedXrangeEntryResponse(entryID string, fieldValues ...string) string {
	var builder strings.Builder

	builder.WriteString("*2\r\n")
	builder.WriteString(formatRESPBulkString(entryID))
	builder.WriteString(fmt.Sprintf("*%d\r\n", len(fieldValues)))
	for _, fieldValue := range fieldValues {
		builder.WriteString(formatRESPBulkString(fieldValue))
	}

	return builder.String()
}

func formatExpectedXrangeResponse(entries ...string) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("*%d\r\n", len(entries)))
	for _, entry := range entries {
		builder.WriteString(entry)
	}

	return builder.String()
}

func TestHandleXrange(t *testing.T) {
	tests := []struct {
		name     string
		setup    func()
		cmd      *RedisCommand
		expected string
	}{
		{
			name: "codecrafters stage returns inclusive range of stream entries",
			setup: func() {
				HandleXadd(&RedisCommand{
					Type: CmdXADD,
					Args: []string{"stream_key", "0-1", "foo", "bar"},
				})
				HandleXadd(&RedisCommand{
					Type: CmdXADD,
					Args: []string{"stream_key", "0-2", "bar", "baz"},
				})
				HandleXadd(&RedisCommand{
					Type: CmdXADD,
					Args: []string{"stream_key", "0-3", "baz", "foo"},
				})
			},
			cmd: &RedisCommand{
				Type: CmdXRANGE,
				Args: []string{"stream_key", "0-2", "0-3"},
			},
			expected: formatExpectedXrangeResponse(
				formatExpectedXrangeEntryResponse("0-2", "bar", "baz"),
				formatExpectedXrangeEntryResponse("0-3", "baz", "foo"),
			),
		},
		{
			name: "returns entries for millisecond-only start and end ids",
			setup: func() {
				HandleXadd(&RedisCommand{
					Type: CmdXADD,
					Args: []string{"some_key", "1526985054069-0", "temperature", "36", "humidity", "95"},
				})
				HandleXadd(&RedisCommand{
					Type: CmdXADD,
					Args: []string{"some_key", "1526985054079-0", "temperature", "37", "humidity", "94"},
				})
			},
			cmd: &RedisCommand{
				Type: CmdXRANGE,
				Args: []string{"some_key", "1526985054069", "1526985054079"},
			},
			expected: formatExpectedXrangeResponse(
				formatExpectedXrangeEntryResponse("1526985054069-0", "temperature", "36", "humidity", "95"),
				formatExpectedXrangeEntryResponse("1526985054079-0", "temperature", "37", "humidity", "94"),
			),
		},
		{
			name: "returns empty array when stream does not exist",
			setup: func() {
			},
			cmd: &RedisCommand{
				Type: CmdXRANGE,
				Args: []string{"missing_stream", "0-1", "0-9"},
			},
			expected: "*0\r\n",
		},
		{
			name: "returns empty array when no entries fall within range",
			setup: func() {
				GetInstance().AddStreamEntry("stream_key", "0-1", []string{"foo", "bar"})
				GetInstance().AddStreamEntry("stream_key", "0-4", []string{"bar", "baz"})
			},
			cmd: &RedisCommand{
				Type: CmdXRANGE,
				Args: []string{"stream_key", "0-2", "0-3"},
			},
			expected: "*0\r\n",
		},
		{
			name: "returns single entry when start and end id match",
			setup: func() {
				GetInstance().AddStreamEntry("stream_key", "0-2", []string{"bar", "baz"})
			},
			cmd: &RedisCommand{
				Type: CmdXRANGE,
				Args: []string{"stream_key", "0-2", "0-2"},
			},
			expected: formatExpectedXrangeResponse(
				formatExpectedXrangeEntryResponse("0-2", "bar", "baz"),
			),
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			GetInstance().cache = make(map[string]CacheItem)

			if testCase.setup != nil {
				testCase.setup()
			}

			result := HandleXrange(testCase.cmd)
			if result != testCase.expected {
				t.Errorf("HandleXrange() = %q, expected %q", result, testCase.expected)
			}
		})
	}
}
