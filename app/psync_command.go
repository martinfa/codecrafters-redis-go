package main

import "fmt"

// HandlePsync processes the PSYNC command and returns a FULLRESYNC response
func HandlePsync(cmd *RedisCommand) string {
	config := GetConfig()
	// The master responds with +FULLRESYNC <REPL_ID> <OFFSET>\r\n
	// For now, we always start with offset 0 as per instructions
	return fmt.Sprintf("+FULLRESYNC %s %d\r\n", config.MasterReplId, config.MasterReplOffset)
}
