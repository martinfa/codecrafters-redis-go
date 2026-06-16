package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const blockingBlpopTimeoutResponse = "*-1\r\n"

func parseBlpopCommandArguments(command *RedisCommand) (listKey string, timeoutSeconds string, errorResponse string) {
	if len(command.Args) != 2 {
		return "", "", "-ERR wrong number of arguments for 'blpop' command\r\n"
	}

	return command.Args[0], command.Args[1], ""
}

func encodeBlpopResponse(listKey string, element string) string {
	var builder strings.Builder

	builder.WriteString("*2\r\n")
	builder.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(listKey), listKey))
	builder.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(element), element))

	return builder.String()
}

func tryImmediateBlpop(listKey string) (string, bool) {
	poppedElements, popped := GetInstance().PopListLeft(listKey, 1)
	if !popped {
		return "", false
	}

	return encodeBlpopResponse(listKey, poppedElements[0]), true
}

func notifyBlockingBlpopWaiters(listKey string) {
	registry := GetBlockingBlpopRegistry()

	for registry.HasWaiters(listKey) {
		poppedElements, popped := GetInstance().PopListLeft(listKey, 1)
		if !popped {
			return
		}

		response := encodeBlpopResponse(listKey, poppedElements[0])
		notified := registry.NotifyNextWaiter(listKey, response)
		if !notified {
			GetInstance().PushListLeft(listKey, poppedElements[0])
			return
		}
	}
}

func waitForBlockingBlpopResponse(listKey string, timeoutSeconds float64) string {
	waiter := GetBlockingBlpopRegistry().RegisterWaiter(listKey)
	defer waiter.Unregister()

	if timeoutSeconds == 0 {
		return <-waiter.ResponseChannel
	}

	timer := time.NewTimer(time.Duration(timeoutSeconds * float64(time.Second)))
	defer timer.Stop()

	select {
	case response := <-waiter.ResponseChannel:
		return response
	case <-timer.C:
		return blockingBlpopTimeoutResponse
	}
}

func HandleBlpop(command *RedisCommand) string {
	listKey, timeoutSecondsString, errorResponse := parseBlpopCommandArguments(command)
	if errorResponse != "" {
		return errorResponse
	}

	timeoutSeconds, parseTimeoutError := strconv.ParseFloat(timeoutSecondsString, 64)
	if parseTimeoutError != nil {
		return "-ERR value is not an integer or out of range\r\n"
	}

	if response, poppedImmediately := tryImmediateBlpop(listKey); poppedImmediately {
		return response
	}

	return waitForBlockingBlpopResponse(listKey, timeoutSeconds)
}
