package main

import (
	"fmt"
	"strconv"
	"strings"
)

func parseLrangeCommandArguments(command *RedisCommand) (listKey string, startIndex string, stopIndex string, errorResponse string) {
	if len(command.Args) != 3 {
		return "", "", "", "-ERR wrong number of arguments for 'lrange' command\r\n"
	}

	return command.Args[0], command.Args[1], command.Args[2], ""
}

func HandleLrange(command *RedisCommand) string {
	listKey, startIndexString, stopIndexString, errorResponse := parseLrangeCommandArguments(command)
	if errorResponse != "" {
		return errorResponse
	}

	startIndex, parseStartError := strconv.Atoi(startIndexString)
	if parseStartError != nil {
		return "-ERR value is not an integer or out of range\r\n"
	}

	stopIndex, parseStopError := strconv.Atoi(stopIndexString)
	if parseStopError != nil {
		return "-ERR value is not an integer or out of range\r\n"
	}

	list := GetInstance().GetList(listKey)
	if list == nil {
		return "*0\r\n"
	}

	listLength := len(list.Elements)
	if startIndex >= listLength {
		return "*0\r\n"
	}

	if stopIndex >= listLength {
		stopIndex = listLength - 1
	}

	if startIndex > stopIndex {
		return "*0\r\n"
	}

	selectedElements := list.Elements[startIndex : stopIndex+1]

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("*%d\r\n", len(selectedElements)))
	for _, element := range selectedElements {
		builder.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(element), element))
	}

	return builder.String()
}
