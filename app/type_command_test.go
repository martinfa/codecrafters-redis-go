package main

import "testing"

func TestHandleType(t *testing.T) {
	tests := []struct {
		name     string
		setup    func()
		cmd      *RedisCommand
		expected string
	}{
		{
			name: "type for existing string key",
			setup: func() {
				Set("some_key", "foo", map[string]interface{}{})
			},
			cmd: &RedisCommand{
				Type: CmdTYPE,
				Args: []string{"some_key"},
			},
			expected: "+string\r\n",
		},
		{
			name: "type for missing key",
			setup: func() {
			},
			cmd: &RedisCommand{
				Type: CmdTYPE,
				Args: []string{"missing_key"},
			},
			expected: "+none\r\n",
		},
		{
			name: "type with no arguments",
			setup: func() {
			},
			cmd: &RedisCommand{
				Type: CmdTYPE,
				Args: []string{},
			},
			expected: "-ERR wrong number of arguments for 'type' command\r\n",
		},
		{
			name: "type with too many arguments",
			setup: func() {
			},
			cmd: &RedisCommand{
				Type: CmdTYPE,
				Args: []string{"key1", "key2"},
			},
			expected: "-ERR wrong number of arguments for 'type' command\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			GetInstance().cache = make(map[string]CacheItem)

			if tt.setup != nil {
				tt.setup()
			}

			result := HandleType(tt.cmd)
			if result != tt.expected {
				t.Errorf("HandleType() = %q, expected %q", result, tt.expected)
			}
		})
	}
}
