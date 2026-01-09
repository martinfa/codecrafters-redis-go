package main

import "fmt"

// HandleInfo processes an INFO command and returns a RESP bulk string response
func HandleInfo(cmd *RedisCommand) string {
	// INFO command can have 0 or 1 argument (section name)
	// For this stage, we only support the replication section
	// The response should be "role:master" or "role:slave" based on configuration

	config := GetConfig()
	role := "master"
	if config.IsReplica {
		role = "slave"
	}
	response := fmt.Sprintf("role:%s", role)

	// Format as RESP bulk string: $<length>\r\n<data>\r\n
	return fmt.Sprintf("$%d\r\n%s\r\n", len(response), response)
}
