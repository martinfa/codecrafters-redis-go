package main

import "fmt"

func parseZcardCommandArguments(command *RedisCommand) (key string, errorResponse string) {
	if len(command.Args) != 1 {
		return "", "-ERR wrong number of arguments for 'zcard' command\r\n"
	}

	return command.Args[0], ""
}

func HandleZcard(command *RedisCommand) string {
	key, errorResponse := parseZcardCommandArguments(command)
	if errorResponse != "" {
		return errorResponse
	}

	memberCount := GetInstance().Zcard(key)
	return fmt.Sprintf(":%d\r\n", memberCount)
}
