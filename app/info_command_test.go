package main

import "testing"

func TestHandleInfo(t *testing.T) {
	tests := []struct {
		name     string
		cmd      *RedisCommand
		expected string
	}{
		{
			name: "info command without arguments",
			cmd: &RedisCommand{
				Type: CmdINFO,
				Args: []string{},
			},
			expected: "$11\r\nrole:master\r\n",
		},
		{
			name: "info command with replication argument",
			cmd: &RedisCommand{
				Type: CmdINFO,
				Args: []string{"replication"},
			},
			expected: "$11\r\nrole:master\r\n",
		},
		{
			name: "info command with multiple arguments",
			cmd: &RedisCommand{
				Type: CmdINFO,
				Args: []string{"replication", "extra"},
			},
			expected: "$11\r\nrole:master\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HandleInfo(tt.cmd)
			if result != tt.expected {
				t.Errorf("HandleInfo() = %q, expected %q", result, tt.expected)
			}
		})
	}
}
