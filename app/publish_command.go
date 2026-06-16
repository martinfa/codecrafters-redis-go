package main

import "fmt"

func parsePublishCommandArguments(command *RedisCommand) (channel string, message string, errorResponse string) {
	if len(command.Args) != 2 {
		return "", "", "-ERR wrong number of arguments for 'publish' command\r\n"
	}

	return command.Args[0], command.Args[1], ""
}

func encodePubSubMessageResponse(channel string, message string) string {
	return fmt.Sprintf(
		"*3\r\n$7\r\nmessage\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n",
		len(channel),
		channel,
		len(message),
		message,
	)
}

func deliverPublishedMessage(channel string, message string) int {
	connectionPubSubMutex.Lock()
	defer connectionPubSubMutex.Unlock()

	encodedMessage := encodePubSubMessageResponse(channel, message)
	subscriberCount := 0

	for connection, pubSubState := range connectionPubSubStates {
		if _, subscribed := pubSubState.subscribedChannels[channel]; !subscribed {
			continue
		}

		subscriberCount++
		writeError := WriteToConnection(connection, encodedMessage)
		if writeError != nil {
			fmt.Printf("Error delivering message to connection: %s\n", writeError.Error())
		}
	}

	return subscriberCount
}

func HandlePublish(command *RedisCommand) string {
	channel, message, errorResponse := parsePublishCommandArguments(command)
	if errorResponse != "" {
		return errorResponse
	}

	subscriberCount := deliverPublishedMessage(channel, message)

	return fmt.Sprintf(":%d\r\n", subscriberCount)
}
