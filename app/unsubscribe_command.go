package main

import (
	"fmt"
	"net"
)

func parseUnsubscribeCommandArguments(command *RedisCommand) (channel string, errorResponse string) {
	if len(command.Args) != 1 {
		return "", "-ERR wrong number of arguments for 'unsubscribe' command\r\n"
	}

	return command.Args[0], ""
}

func encodeUnsubscribeResponse(channel string, remainingSubscriptionCount int) string {
	return fmt.Sprintf(
		"*3\r\n$11\r\nunsubscribe\r\n$%d\r\n%s\r\n:%d\r\n",
		len(channel),
		channel,
		remainingSubscriptionCount,
	)
}

func HandleUnsubscribe(connection net.Conn, command *RedisCommand) string {
	channel, errorResponse := parseUnsubscribeCommandArguments(command)
	if errorResponse != "" {
		return errorResponse
	}

	pubSubState := getConnectionPubSubState(connection)
	if _, subscribed := pubSubState.subscribedChannels[channel]; subscribed {
		delete(pubSubState.subscribedChannels, channel)
	}

	return encodeUnsubscribeResponse(channel, len(pubSubState.subscribedChannels))
}
