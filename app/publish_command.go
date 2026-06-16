package main

import "fmt"

func parsePublishCommandArguments(command *RedisCommand) (channel string, message string, errorResponse string) {
	if len(command.Args) != 2 {
		return "", "", "-ERR wrong number of arguments for 'publish' command\r\n"
	}

	return command.Args[0], command.Args[1], ""
}

func countChannelSubscribers(channel string) int {
	connectionPubSubMutex.Lock()
	defer connectionPubSubMutex.Unlock()

	subscriberCount := 0
	for _, pubSubState := range connectionPubSubStates {
		if _, subscribed := pubSubState.subscribedChannels[channel]; subscribed {
			subscriberCount++
		}
	}

	return subscriberCount
}

func HandlePublish(command *RedisCommand) string {
	channel, _, errorResponse := parsePublishCommandArguments(command)
	if errorResponse != "" {
		return errorResponse
	}

	subscriberCount := countChannelSubscribers(channel)

	return fmt.Sprintf(":%d\r\n", subscriberCount)
}
