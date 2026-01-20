package main

import (
	"encoding/hex"
	"fmt"
)

// HandlePsync processes the PSYNC command and returns a FULLRESYNC response followed by an empty RDB file
func HandlePsync(cmd *RedisCommand) string {
	config := GetConfig()
	// 1. The master responds with +FULLRESYNC <REPL_ID> <OFFSET>\r\n
	fullResyncResp := fmt.Sprintf("+FULLRESYNC %s %d\r\n", config.MasterReplId, config.MasterReplOffset)

	// 2. The master sends an empty RDB file
	// Hex for a minimal empty RDB file
	rdbHex := "524544495330303131fa0972656469732d76657205372e322e30fa0a72656469732d62697473c040fa056374696d65c26d08bc65fa08757365642d6d656dc2b0c41000fa08616f662d62617365c000ff10aa32556141e212"
	rdbBytes, _ := hex.DecodeString(rdbHex)

	// Format: $<length>\r\n<contents> (no trailing \r\n)
	rdbResp := fmt.Sprintf("$%d\r\n%s", len(rdbBytes), string(rdbBytes))

	return fullResyncResp + rdbResp
}
