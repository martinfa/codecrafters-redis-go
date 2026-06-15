package main

import (
	"fmt"
	"strconv"
	"strings"
)

func parseLpopCommandArguments(command *RedisCommand) (listKey string, popCount string, hasPopCount bool, errorResponse string) {
	if len(command.Args) < 1 || len(command.Args) > 2 {
		return "", "", false, "-ERR wrong number of arguments for 'lpop' command\r\n"
	}

	if len(command.Args) == 1 {
		return command.Args[0], "", false, ""
	}

	return command.Args[0], command.Args[1], true, ""
}

func encodeLpopArrayResponse(elements []string) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("*%d\r\n", len(elements)))
	for _, element := range elements {
		builder.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(element), element))
	}

	return builder.String()
}

func HandleLpop(command *RedisCommand) string {
	listKey, popCountString, hasPopCount, errorResponse := parseLpopCommandArguments(command)
	if errorResponse != "" {
		return errorResponse
	}

	popCount := 1
	if hasPopCount {
		parsedPopCount, parsePopCountError := strconv.Atoi(popCountString)
		if parsePopCountError != nil {
			return "-ERR value is not an integer or out of range\r\n"
		}

		popCount = parsedPopCount
	}

	poppedElements, popped := GetInstance().PopListLeft(listKey, popCount)
	if !popped {
		return "$-1\r\n"
	}

	if !hasPopCount {
		return fmt.Sprintf("$%d\r\n%s\r\n", len(poppedElements[0]), poppedElements[0])
	}

	return encodeLpopArrayResponse(poppedElements)
}
