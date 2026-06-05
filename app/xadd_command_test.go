package main

import "testing"

func TestHandleXadd(t *testing.T) {
	tests := []struct {
		name     string
		setup    func()
		cmd      *RedisCommand
		expected string
	}{
		{
			name: "xadd creates stream and returns entry id",
			setup: func() {
			},
			cmd: &RedisCommand{
				Type: CmdXADD,
				Args: []string{"stream_key", "0-1", "foo", "bar"},
			},
			expected: "$3\r\n0-1\r\n",
		},
		{
			name: "xadd appends to existing stream",
			setup: func() {
				GetInstance().AddStreamEntry("stream_key", "0-1", []string{"foo", "bar"})
			},
			cmd: &RedisCommand{
				Type: CmdXADD,
				Args: []string{"stream_key", "0-2", "temperature", "36", "humidity", "95"},
			},
			expected: "$3\r\n0-2\r\n",
		},
		{
			name: "xadd with too few arguments",
			setup: func() {
			},
			cmd: &RedisCommand{
				Type: CmdXADD,
				Args: []string{"stream_key", "0-1"},
			},
			expected: "-ERR wrong number of arguments for 'xadd' command\r\n",
		},
		{
			name: "xadd with odd number of field arguments",
			setup: func() {
			},
			cmd: &RedisCommand{
				Type: CmdXADD,
				Args: []string{"stream_key", "0-1", "foo"},
			},
			expected: "-ERR wrong number of arguments for 'xadd' command\r\n",
		},
		{
			name: "xadd rejects id equal to last entry",
			setup: func() {
				GetInstance().AddStreamEntry("stream_key", "1-1", []string{"foo", "bar"})
				GetInstance().AddStreamEntry("stream_key", "1-2", []string{"bar", "baz"})
			},
			cmd: &RedisCommand{
				Type: CmdXADD,
				Args: []string{"stream_key", "1-2", "baz", "foo"},
			},
			expected: errXaddIDEqualOrSmallerThanTop,
		},
		{
			name: "xadd rejects id with lower milliseconds than last entry",
			setup: func() {
				GetInstance().AddStreamEntry("stream_key", "1-1", []string{"foo", "bar"})
				GetInstance().AddStreamEntry("stream_key", "1-2", []string{"bar", "baz"})
			},
			cmd: &RedisCommand{
				Type: CmdXADD,
				Args: []string{"stream_key", "0-3", "baz", "foo"},
			},
			expected: errXaddIDEqualOrSmallerThanTop,
		},
		{
			name: "xadd rejects 0-0 on empty stream",
			setup: func() {
			},
			cmd: &RedisCommand{
				Type: CmdXADD,
				Args: []string{"stream_key", "0-0", "baz", "foo"},
			},
			expected: errXaddIDMustBeGreaterThanZeroZero,
		},
		{
			name: "xadd rejects 0-0 when stream has entries",
			setup: func() {
				GetInstance().AddStreamEntry("stream_key", "1-1", []string{"foo", "bar"})
			},
			cmd: &RedisCommand{
				Type: CmdXADD,
				Args: []string{"stream_key", "0-0", "baz", "foo"},
			},
			expected: errXaddIDMustBeGreaterThanZeroZero,
		},
		{
			name: "xadd rejects duplicate id on single entry stream",
			setup: func() {
				GetInstance().AddStreamEntry("stream_key", "1-1", []string{"foo", "bar"})
			},
			cmd: &RedisCommand{
				Type: CmdXADD,
				Args: []string{"stream_key", "1-1", "bar", "baz"},
			},
			expected: errXaddIDEqualOrSmallerThanTop,
		},
		{
			name: "xadd auto generates sequence for 0-* on empty stream",
			setup: func() {
			},
			cmd: &RedisCommand{
				Type: CmdXADD,
				Args: []string{"stream_key", "0-*", "foo", "bar"},
			},
			expected: "$3\r\n0-1\r\n",
		},
		{
			name: "xadd auto generates first sequence for new milliseconds",
			setup: func() {
				GetInstance().AddStreamEntry("stream_key", "0-1", []string{"foo", "bar"})
			},
			cmd: &RedisCommand{
				Type: CmdXADD,
				Args: []string{"stream_key", "5-*", "foo", "bar"},
			},
			expected: "$3\r\n5-0\r\n",
		},
		{
			name: "xadd auto generates incremented sequence for same milliseconds",
			setup: func() {
				GetInstance().AddStreamEntry("stream_key", "5-0", []string{"foo", "bar"})
			},
			cmd: &RedisCommand{
				Type: CmdXADD,
				Args: []string{"stream_key", "5-*", "bar", "baz"},
			},
			expected: "$3\r\n5-1\r\n",
		},
		{
			name: "xadd auto generates incremented sequence for same milliseconds after other time parts",
			setup: func() {
				GetInstance().AddStreamEntry("stream_key", "1-0", []string{"a", "1"})
				GetInstance().AddStreamEntry("stream_key", "1-5", []string{"b", "2"})
			},
			cmd: &RedisCommand{
				Type: CmdXADD,
				Args: []string{"stream_key", "1-*", "c", "3"},
			},
			expected: "$3\r\n1-6\r\n",
		},
		{
			name: "xadd auto generates sequence for * on empty stream",
			setup: func() {
				CurrentTimeMilliseconds = func() int64 {
					return 1526919030474
				}
			},
			cmd: &RedisCommand{
				Type: CmdXADD,
				Args: []string{"stream_key", "*", "foo", "bar"},
			},
			expected: "$15\r\n1526919030474-0\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalCurrentTimeMilliseconds := CurrentTimeMilliseconds
			t.Cleanup(func() {
				CurrentTimeMilliseconds = originalCurrentTimeMilliseconds
			})

			GetInstance().cache = make(map[string]CacheItem)

			if tt.setup != nil {
				tt.setup()
			}

			result := HandleXadd(tt.cmd)
			if result != tt.expected {
				t.Errorf("HandleXadd() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestHandleXaddStoresFieldValues(t *testing.T) {
	GetInstance().cache = make(map[string]CacheItem)

	HandleXadd(&RedisCommand{
		Type: CmdXADD,
		Args: []string{"stream_key", "1526919030474-0", "temperature", "36", "humidity", "95"},
	})

	stream := GetInstance().GetStream("stream_key")
	if stream == nil {
		t.Fatal("expected stream to exist")
	}

	if len(stream.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(stream.Entries))
	}

	entry := stream.Entries[0]
	if entry.ID != "1526919030474-0" {
		t.Errorf("expected entry ID %q, got %q", "1526919030474-0", entry.ID)
	}

	if entry.Fields["temperature"] != "36" {
		t.Errorf("expected temperature 36, got %q", entry.Fields["temperature"])
	}

	if entry.Fields["humidity"] != "95" {
		t.Errorf("expected humidity 95, got %q", entry.Fields["humidity"])
	}
}
