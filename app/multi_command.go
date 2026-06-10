package main

func parseMultiCommandArguments(command *RedisCommand) (errorResponse string) {
	if len(command.Args) != 0 {
		return "-ERR wrong number of arguments for 'multi' command\r\n"
	}

	return ""
}

func HandleMulti(command *RedisCommand) string {
	errorResponse := parseMultiCommandArguments(command)
	if errorResponse != "" {
		return errorResponse
	}

	return "+OK\r\n"
}
