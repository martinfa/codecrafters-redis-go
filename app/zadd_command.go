package main

import (
	"fmt"
	"strconv"
)

func parseZaddCommandArguments(command *RedisCommand) (key string, score float64, member string, errorResponse string) {
	if len(command.Args) != 3 {
		return "", 0, "", "-ERR wrong number of arguments for 'zadd' command\r\n"
	}

	parsedScore, parseError := strconv.ParseFloat(command.Args[1], 64)
	if parseError != nil {
		return "", 0, "", "-ERR value is not a valid float\r\n"
	}

	return command.Args[0], parsedScore, command.Args[2], ""
}

func HandleZadd(command *RedisCommand) string {
	key, score, member, errorResponse := parseZaddCommandArguments(command)
	if errorResponse != "" {
		return errorResponse
	}

	newMembersAdded := GetInstance().Zadd(key, score, member)
	return fmt.Sprintf(":%d\r\n", newMembersAdded)
}
