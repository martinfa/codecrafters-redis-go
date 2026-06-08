package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestParseXreadCommandArguments(t *testing.T) {
	tests := []struct {
		name              string
		args              []string
		expectedStreamKey string
		expectedStartID   string
		expectedError     string
	}{
		{
			name:              "parses streams keyword stream key and start id",
			args:              []string{"STREAMS", "stream_key", "0-0"},
			expectedStreamKey: "stream_key",
			expectedStartID:   "0-0",
		},
		{
			name:              "parses lowercase streams keyword",
			args:              []string{"streams", "some_key", "1526985054069-0"},
			expectedStreamKey: "some_key",
			expectedStartID:   "1526985054069-0",
		},
		{
			name:          "rejects no arguments",
			args:          []string{},
			expectedError: "-ERR wrong number of arguments for 'xread' command\r\n",
		},
		{
			name:          "rejects missing start id",
			args:          []string{"STREAMS", "stream_key"},
			expectedError: "-ERR wrong number of arguments for 'xread' command\r\n",
		},
		{
			name:          "rejects missing streams keyword",
			args:          []string{"stream_key", "0-0"},
			expectedError: "-ERR wrong number of arguments for 'xread' command\r\n",
		},
		{
			name:          "rejects invalid subcommand",
			args:          []string{"COUNT", "stream_key", "0-0"},
			expectedError: "-ERR wrong number of arguments for 'xread' command\r\n",
		},
		{
			name:          "rejects too many arguments",
			args:          []string{"STREAMS", "stream_key", "0-0", "extra"},
			expectedError: "-ERR wrong number of arguments for 'xread' command\r\n",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			command := &RedisCommand{
				Type: CmdXREAD,
				Args: testCase.args,
			}

			streamKey, startID, errorResponse := parseXreadCommandArguments(command)

			if errorResponse != testCase.expectedError {
				t.Errorf("parseXreadCommandArguments() error = %q, expected %q", errorResponse, testCase.expectedError)
			}

			if streamKey != testCase.expectedStreamKey {
				t.Errorf("parseXreadCommandArguments() streamKey = %q, expected %q", streamKey, testCase.expectedStreamKey)
			}

			if startID != testCase.expectedStartID {
				t.Errorf("parseXreadCommandArguments() startID = %q, expected %q", startID, testCase.expectedStartID)
			}
		})
	}
}

func TestHandleXreadArgumentParsing(t *testing.T) {
	tests := []struct {
		name     string
		cmd      *RedisCommand
		expected string
	}{
		{
			name: "invalid argument count returns error",
			cmd: &RedisCommand{
				Type: CmdXREAD,
				Args: []string{"STREAMS", "stream_key"},
			},
			expected: "-ERR wrong number of arguments for 'xread' command\r\n",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := HandleXread(testCase.cmd)
			if result != testCase.expected {
				t.Errorf("HandleXread() = %q, expected %q", result, testCase.expected)
			}
		})
	}
}

func formatExpectedXreadEntryResponse(entryID string, fieldValues ...string) string {
	var builder strings.Builder

	builder.WriteString("*2\r\n")
	builder.WriteString(formatRESPBulkString(entryID))
	builder.WriteString(fmt.Sprintf("*%d\r\n", len(fieldValues)))
	for _, fieldValue := range fieldValues {
		builder.WriteString(formatRESPBulkString(fieldValue))
	}

	return builder.String()
}

func formatExpectedXreadStreamResponse(streamKey string, entries ...string) string {
	var builder strings.Builder

	builder.WriteString("*2\r\n")
	builder.WriteString(formatRESPBulkString(streamKey))
	builder.WriteString(fmt.Sprintf("*%d\r\n", len(entries)))
	for _, entry := range entries {
		builder.WriteString(entry)
	}

	return builder.String()
}

func formatExpectedXreadResponse(streams ...string) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("*%d\r\n", len(streams)))
	for _, stream := range streams {
		builder.WriteString(stream)
	}

	return builder.String()
}

func TestHandleXread(t *testing.T) {
	tests := []struct {
		name     string
		setup    func()
		cmd      *RedisCommand
		expected string
	}{
		{
			name: "codecrafters stage returns entries with id greater than start id",
			setup: func() {
				HandleXadd(&RedisCommand{
					Type: CmdXADD,
					Args: []string{"stream_key", "0-1", "temperature", "96"},
				})
			},
			cmd: &RedisCommand{
				Type: CmdXREAD,
				Args: []string{"STREAMS", "stream_key", "0-0"},
			},
			expected: formatExpectedXreadResponse(
				formatExpectedXreadStreamResponse(
					"stream_key",
					formatExpectedXreadEntryResponse("0-1", "temperature", "96"),
				),
			),
		},
		{
			name: "returns only entries strictly after the provided start id",
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
				Type: CmdXREAD,
				Args: []string{"STREAMS", "some_key", "1526985054069-0"},
			},
			expected: formatExpectedXreadResponse(
				formatExpectedXreadStreamResponse(
					"some_key",
					formatExpectedXreadEntryResponse("1526985054079-0", "temperature", "37", "humidity", "94"),
				),
			),
		},
		{
			name: "returns empty array when no entries are after start id",
			setup: func() {
				GetInstance().AddStreamEntry("stream_key", "0-1", []string{"temperature", "96"})
			},
			cmd: &RedisCommand{
				Type: CmdXREAD,
				Args: []string{"STREAMS", "stream_key", "0-1"},
			},
			expected: "*0\r\n",
		},
		{
			name: "returns empty array when stream does not exist",
			setup: func() {
			},
			cmd: &RedisCommand{
				Type: CmdXREAD,
				Args: []string{"STREAMS", "missing_stream", "0-0"},
			},
			expected: "*0\r\n",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			GetInstance().cache = make(map[string]CacheItem)

			if testCase.setup != nil {
				testCase.setup()
			}

			result := HandleXread(testCase.cmd)
			if result != testCase.expected {
				t.Errorf("HandleXread() = %q, expected %q", result, testCase.expected)
			}
		})
	}
}

func TestHandleXreadMultipleStreams(t *testing.T) {
	tests := []struct {
		name     string
		setup    func()
		cmd      *RedisCommand
		expected string
	}{
		{
			name: "codecrafters stage returns multiple streams in command order",
			setup: func() {
				HandleXadd(&RedisCommand{
					Type: CmdXADD,
					Args: []string{"stream_key", "0-1", "temperature", "95"},
				})
				HandleXadd(&RedisCommand{
					Type: CmdXADD,
					Args: []string{"other_stream_key", "0-2", "humidity", "97"},
				})
			},
			cmd: &RedisCommand{
				Type: CmdXREAD,
				Args: []string{"streams", "stream_key", "other_stream_key", "0-0", "0-1"},
			},
			expected: formatExpectedXreadResponse(
				formatExpectedXreadStreamResponse(
					"stream_key",
					formatExpectedXreadEntryResponse("0-1", "temperature", "95"),
				),
				formatExpectedXreadStreamResponse(
					"other_stream_key",
					formatExpectedXreadEntryResponse("0-2", "humidity", "97"),
				),
			),
		},
		{
			name: "returns streams in the same order as the command keys",
			setup: func() {
				HandleXadd(&RedisCommand{
					Type: CmdXADD,
					Args: []string{"first_stream", "0-1", "foo", "bar"},
				})
				HandleXadd(&RedisCommand{
					Type: CmdXADD,
					Args: []string{"second_stream", "0-1", "bar", "baz"},
				})
			},
			cmd: &RedisCommand{
				Type: CmdXREAD,
				Args: []string{"STREAMS", "second_stream", "first_stream", "0-0", "0-0"},
			},
			expected: formatExpectedXreadResponse(
				formatExpectedXreadStreamResponse(
					"second_stream",
					formatExpectedXreadEntryResponse("0-1", "bar", "baz"),
				),
				formatExpectedXreadStreamResponse(
					"first_stream",
					formatExpectedXreadEntryResponse("0-1", "foo", "bar"),
				),
			),
		},
		{
			name: "returns empty array when no stream has entries after start id",
			setup: func() {
				GetInstance().AddStreamEntry("stream_key", "0-1", []string{"temperature", "95"})
				GetInstance().AddStreamEntry("other_stream_key", "0-2", []string{"humidity", "97"})
			},
			cmd: &RedisCommand{
				Type: CmdXREAD,
				Args: []string{"STREAMS", "stream_key", "other_stream_key", "0-1", "0-2"},
			},
			expected: "*0\r\n",
		},
		{
			name: "rejects mismatched stream and id counts",
			cmd: &RedisCommand{
				Type: CmdXREAD,
				Args: []string{"STREAMS", "stream_key", "other_stream_key", "0-0"},
			},
			expected: "-ERR wrong number of arguments for 'xread' command\r\n",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			GetInstance().cache = make(map[string]CacheItem)

			if testCase.setup != nil {
				testCase.setup()
			}

			result := HandleXread(testCase.cmd)
			if result != testCase.expected {
				t.Errorf("HandleXread() = %q, expected %q", result, testCase.expected)
			}
		})
	}
}
