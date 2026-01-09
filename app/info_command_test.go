package main

import "testing"

func TestHandleInfo(t *testing.T) {
	// Save original config
	originalConfig := serverConfig
	defer func() { serverConfig = originalConfig }() // Restore config after test

	tests := []struct {
		name     string
		setup    func()
		cmd      *RedisCommand
		expected string
	}{
		{
			name: "info command without arguments - master",
			setup: func() {
				serverConfig.IsReplica = false
			},
			cmd: &RedisCommand{
				Type: CmdINFO,
				Args: []string{},
			},
			expected: "$11\r\nrole:master\r\n",
		},
		{
			name: "info command with replication argument - master",
			setup: func() {
				serverConfig.IsReplica = false
			},
			cmd: &RedisCommand{
				Type: CmdINFO,
				Args: []string{"replication"},
			},
			expected: "$11\r\nrole:master\r\n",
		},
		{
			name: "info command with multiple arguments - master",
			setup: func() {
				serverConfig.IsReplica = false
			},
			cmd: &RedisCommand{
				Type: CmdINFO,
				Args: []string{"replication", "extra"},
			},
			expected: "$11\r\nrole:master\r\n",
		},
		{
			name: "info command without arguments - replica",
			setup: func() {
				serverConfig.IsReplica = true
			},
			cmd: &RedisCommand{
				Type: CmdINFO,
				Args: []string{},
			},
			expected: "$10\r\nrole:slave\r\n",
		},
		{
			name: "info command with replication argument - replica",
			setup: func() {
				serverConfig.IsReplica = true
			},
			cmd: &RedisCommand{
				Type: CmdINFO,
				Args: []string{"replication"},
			},
			expected: "$10\r\nrole:slave\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			result := HandleInfo(tt.cmd)
			if result != tt.expected {
				t.Errorf("HandleInfo() = %q, expected %q", result, tt.expected)
			}
		})
	}
}
