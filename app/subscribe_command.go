package main

import (
	"fmt"
	"net"
	"strings"
	"sync"
)

type connectionPubSubState struct {
	subscribedChannels map[string]struct{}
}

var (
	connectionPubSubMutex  sync.Mutex
	connectionPubSubStates = make(map[net.Conn]*connectionPubSubState)
)

func ResetConnectionPubSubStatesForTest() {
	connectionPubSubMutex.Lock()
	defer connectionPubSubMutex.Unlock()

	connectionPubSubStates = make(map[net.Conn]*connectionPubSubState)
}

func getConnectionPubSubState(connection net.Conn) *connectionPubSubState {
	connectionPubSubMutex.Lock()
	defer connectionPubSubMutex.Unlock()

	state, exists := connectionPubSubStates[connection]
	if !exists {
		state = &connectionPubSubState{
			subscribedChannels: make(map[string]struct{}),
		}
		connectionPubSubStates[connection] = state
	}

	return state
}

func RemoveConnectionPubSubState(connection net.Conn) {
	connectionPubSubMutex.Lock()
	defer connectionPubSubMutex.Unlock()

	delete(connectionPubSubStates, connection)
}

func getConnectionPubSubStateIfExists(connection net.Conn) (*connectionPubSubState, bool) {
	connectionPubSubMutex.Lock()
	defer connectionPubSubMutex.Unlock()

	state, exists := connectionPubSubStates[connection]
	return state, exists
}

func isConnectionInSubscribedMode(connection net.Conn) bool {
	state, exists := getConnectionPubSubStateIfExists(connection)
	if !exists {
		return false
	}

	return len(state.subscribedChannels) > 0
}

func isCommandAllowedInSubscribedMode(commandType CommandType) bool {
	switch commandType {
	case CmdSUBSCRIBE, CmdUNSUBSCRIBE, CmdPSUBSCRIBE, CmdPUNSUBSCRIBE, CmdPING, CmdQUIT, CmdRESET:
		return true
	default:
		return false
	}
}

func subscribedModeErrorResponse(commandType CommandType) string {
	commandName := strings.ToLower(commandType.String())

	return fmt.Sprintf(
		"-ERR Can't execute '%s': only (P|S)SUBSCRIBE / (P|S)UNSUBSCRIBE / PING / QUIT / RESET are allowed in this context\r\n",
		commandName,
	)
}

func parseSubscribeCommandArguments(command *RedisCommand) (channel string, errorResponse string) {
	if len(command.Args) != 1 {
		return "", "-ERR wrong number of arguments for 'subscribe' command\r\n"
	}

	return command.Args[0], ""
}

func encodeSubscribeResponse(channel string, subscriptionCount int) string {
	return fmt.Sprintf(
		"*3\r\n$9\r\nsubscribe\r\n$%d\r\n%s\r\n:%d\r\n",
		len(channel),
		channel,
		subscriptionCount,
	)
}

func HandleSubscribe(connection net.Conn, command *RedisCommand) string {
	channel, errorResponse := parseSubscribeCommandArguments(command)
	if errorResponse != "" {
		return errorResponse
	}

	pubSubState := getConnectionPubSubState(connection)
	pubSubState.subscribedChannels[channel] = struct{}{}

	return encodeSubscribeResponse(channel, len(pubSubState.subscribedChannels))
}
