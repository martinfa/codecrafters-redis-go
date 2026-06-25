package main

import (
	"fmt"
	"strconv"
	"strings"
)

func parseZrangeCommandArguments(command *RedisCommand) (key string, startIndex string, stopIndex string, errorResponse string) {
	if len(command.Args) != 3 {
		return "", "", "", "-ERR wrong number of arguments for 'zrange' command\r\n"
	}

	return command.Args[0], command.Args[1], command.Args[2], ""
}

func encodeZrangeResponse(members []string) string {
	if len(members) == 0 {
		return "*0\r\n"
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("*%d\r\n", len(members)))
	for _, member := range members {
		builder.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(member), member))
	}

	return builder.String()
}

func HandleZrange(command *RedisCommand) string {
	key, startIndexString, stopIndexString, errorResponse := parseZrangeCommandArguments(command)
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

	sortedSet := GetInstance().GetSortedSet(key)
	if sortedSet == nil {
		return "*0\r\n"
	}

	sortedSetLength := sortedSet.orderedIndex.Length()

	startIndex = normalizeLrangeIndex(startIndex, sortedSetLength)
	if startIndex >= sortedSetLength {
		return "*0\r\n"
	}

	stopIndex = normalizeLrangeIndex(stopIndex, sortedSetLength)
	if stopIndex >= sortedSetLength {
		stopIndex = sortedSetLength - 1
	}

	if startIndex > stopIndex {
		return "*0\r\n"
	}

	members := GetInstance().Zrange(key, startIndex, stopIndex)
	return encodeZrangeResponse(members)
}
