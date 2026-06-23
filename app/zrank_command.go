package main

import "fmt"

func parseZrankCommandArguments(command *RedisCommand) (key string, member string, errorResponse string) {
	if len(command.Args) != 2 {
		return "", "", "-ERR wrong number of arguments for 'zrank' command\r\n"
	}

	return command.Args[0], command.Args[1], ""
}

func encodeZrankResponse(rank int) string {
	return fmt.Sprintf(":%d\r\n", rank)
}

func encodeZrankNullResponse() string {
	return "$-1\r\n"
}

func HandleZrank(command *RedisCommand) string {
	key, member, errorResponse := parseZrankCommandArguments(command)
	if errorResponse != "" {
		return errorResponse
	}

	rank, found := GetInstance().Zrank(key, member)
	if !found {
		return encodeZrankNullResponse()
	}

	return encodeZrankResponse(rank)
}
