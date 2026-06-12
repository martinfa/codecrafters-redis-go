package main

import "fmt"

func parseLlenCommandArguments(command *RedisCommand) (listKey string, errorResponse string) {
	if len(command.Args) != 1 {
		return "", "-ERR wrong number of arguments for 'llen' command\r\n"
	}

	return command.Args[0], ""
}

func HandleLlen(command *RedisCommand) string {
	listKey, errorResponse := parseLlenCommandArguments(command)
	if errorResponse != "" {
		return errorResponse
	}

	listLength := GetInstance().GetListLength(listKey)

	return fmt.Sprintf(":%d\r\n", listLength)
}
