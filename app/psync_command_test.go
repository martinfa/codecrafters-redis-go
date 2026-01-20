package main

import (
	"fmt"
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

	expected := "+FULLRESYNC test_id 0\r\n"
	result := HandlePsync(cmd)

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
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

	expected := fmt.Sprintf("+FULLRESYNC abc 42\r\n")
	result := HandlePsync(cmd)

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}
