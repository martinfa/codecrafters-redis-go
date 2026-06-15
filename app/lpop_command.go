package main

import "fmt"

func parseLpopCommandArguments(command *RedisCommand) (listKey string, errorResponse string) {
	if len(command.Args) != 1 {
		return "", "-ERR wrong number of arguments for 'lpop' command\r\n"
	}

	return command.Args[0], ""
}

func HandleLpop(command *RedisCommand) string {
	listKey, errorResponse := parseLpopCommandArguments(command)
	if errorResponse != "" {
		return errorResponse
	}

	poppedElement, popped := GetInstance().PopListLeft(listKey)
	if !popped {
		return "$-1\r\n"
	}

	return fmt.Sprintf("$%d\r\n%s\r\n", len(poppedElement), poppedElement)
}
