package main

import "fmt"

func parseLpushCommandArguments(command *RedisCommand) (listKey string, elements []string, errorResponse string) {
	if len(command.Args) < 2 {
		return "", nil, "-ERR wrong number of arguments for 'lpush' command\r\n"
	}

	return command.Args[0], command.Args[1:], ""
}

func HandleLpush(command *RedisCommand) string {
	listKey, elements, errorResponse := parseLpushCommandArguments(command)
	if errorResponse != "" {
		return errorResponse
	}

	listLength := GetInstance().PushListLeft(listKey, elements...)

	return fmt.Sprintf(":%d\r\n", listLength)
}
