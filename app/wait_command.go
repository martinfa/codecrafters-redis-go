package main

import (
	"fmt"
	"strconv"
)

// HandleWait processes a WAIT command and returns a RESP integer response.
func HandleWait(command *RedisCommand) string {
	if len(command.Args) != 2 {
		return "-ERR wrong number of arguments for 'wait' command\r\n"
	}

	if _, parseError := strconv.Atoi(command.Args[0]); parseError != nil {
		return fmt.Sprintf("-ERR invalid numreplicas value %q\r\n", command.Args[0])
	}

	if _, parseError := strconv.Atoi(command.Args[1]); parseError != nil {
		return fmt.Sprintf("-ERR invalid timeout value %q\r\n", command.Args[1])
	}

	return fmt.Sprintf(":%d\r\n", ReplicaCount())
}
