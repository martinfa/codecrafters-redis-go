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
				serverConfig.MasterReplId = "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb"
				serverConfig.MasterReplOffset = 0
			},
			cmd: &RedisCommand{
				Type: CmdINFO,
				Args: []string{},
			},
			expected: "$87\r\nrole:master\nmaster_replid:8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb\nmaster_repl_offset:0\r\n",
		},
		{
			name: "info command with replication argument - master",
			setup: func() {
				serverConfig.IsReplica = false
				serverConfig.MasterReplId = "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb"
				serverConfig.MasterReplOffset = 0
			},
			cmd: &RedisCommand{
				Type: CmdINFO,
				Args: []string{"replication"},
			},
			expected: "$87\r\nrole:master\nmaster_replid:8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb\nmaster_repl_offset:0\r\n",
		},
		{
			name: "info command with multiple arguments - master",
			setup: func() {
				serverConfig.IsReplica = false
				serverConfig.MasterReplId = "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb"
				serverConfig.MasterReplOffset = 0
			},
			cmd: &RedisCommand{
				Type: CmdINFO,
				Args: []string{"replication", "extra"},
			},
			expected: "$87\r\nrole:master\nmaster_replid:8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb\nmaster_repl_offset:0\r\n",
		},
		{
			name: "info command without arguments - replica",
			setup: func() {
				serverConfig.IsReplica = true
				serverConfig.MasterReplId = "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb"
				serverConfig.MasterReplOffset = 0
			},
			cmd: &RedisCommand{
				Type: CmdINFO,
				Args: []string{},
			},
			expected: "$86\r\nrole:slave\nmaster_replid:8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb\nmaster_repl_offset:0\r\n",
		},
		{
			name: "info command with replication argument - replica",
			setup: func() {
				serverConfig.IsReplica = true
				serverConfig.MasterReplId = "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb"
				serverConfig.MasterReplOffset = 0
			},
			cmd: &RedisCommand{
				Type: CmdINFO,
				Args: []string{"replication"},
			},
			expected: "$86\r\nrole:slave\nmaster_replid:8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb\nmaster_repl_offset:0\r\n",
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
