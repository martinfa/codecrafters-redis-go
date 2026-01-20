package main

import (
	"testing"
)

func TestHandlePsync(t *testing.T) {
	// Initialize config for the test
	serverConfig.MasterReplId = "test_id"
	serverConfig.MasterReplOffset = 0

	cmd := &RedisCommand{
		Type: CmdPSYNC,
		Args: []string{"?", "-1"},
	}

	result := HandlePsync(cmd)

	// Check for FULLRESYNC part
	expectedPrefix := "+FULLRESYNC test_id 0\r\n"
	if result[:len(expectedPrefix)] != expectedPrefix {
		t.Errorf("Expected prefix %q, got %q", expectedPrefix, result[:len(expectedPrefix)])
	}

	// Check for RDB part ($<length>\r\nREDIS...)
	if result[len(expectedPrefix):len(expectedPrefix)+1] != "$" {
		t.Errorf("Expected $ at start of RDB part, got %q", result[len(expectedPrefix):len(expectedPrefix)+1])
	}
}

func TestHandlePsync_CustomOffset(t *testing.T) {
	// Initialize config for the test with a different offset
	serverConfig.MasterReplId = "abc"
	serverConfig.MasterReplOffset = 42

	cmd := &RedisCommand{
		Type: CmdPSYNC,
		Args: []string{"?", "-1"},
	}

	result := HandlePsync(cmd)

	expectedPrefix := "+FULLRESYNC abc 42\r\n"
	if result[:len(expectedPrefix)] != expectedPrefix {
		t.Errorf("Expected prefix %q, got %q", expectedPrefix, result[:len(expectedPrefix)])
	}
}
