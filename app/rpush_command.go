package main

import "fmt"

func parseRpushCommandArguments(command *RedisCommand) (listKey string, elements []string, errorResponse string) {
	if len(command.Args) < 2 {
		return "", nil, "-ERR wrong number of arguments for 'rpush' command\r\n"
	}

	return command.Args[0], command.Args[1:], ""
}

func HandleRpush(command *RedisCommand) string {
	listKey, elements, errorResponse := parseRpushCommandArguments(command)
	if errorResponse != "" {
		return errorResponse
	}

	listLength := GetInstance().PushListRight(listKey, elements...)
	notifyBlockingBlpopWaiters(listKey)

	return fmt.Sprintf(":%d\r\n", listLength)
}
