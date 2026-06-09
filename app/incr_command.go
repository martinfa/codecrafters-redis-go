package main

import (
	"fmt"
	"strconv"
)

func parseIncrCommandArguments(command *RedisCommand) (key string, errorResponse string) {
	if len(command.Args) != 1 {
		return "", "-ERR wrong number of arguments for 'incr' command\r\n"
	}

	return command.Args[0], ""
}

func HandleIncr(command *RedisCommand) string {
	key, errorResponse := parseIncrCommandArguments(command)
	if errorResponse != "" {
		return errorResponse
	}

	cache := GetInstance()
	value := cache.Get(key)

	valueString := "0"
	if value != nil {
		valueString = value.(string)
	}

	currentValue, parseError := strconv.Atoi(valueString)
	if parseError != nil {
		return "-ERR value is not an integer or out of range\r\n"
	}

	incrementedValue := currentValue + 1
	cache.Set(key, strconv.Itoa(incrementedValue), nil)

	return fmt.Sprintf(":%d\r\n", incrementedValue)
}
