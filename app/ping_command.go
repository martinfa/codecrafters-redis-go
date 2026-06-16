package main

import "net"

const normalPingResponse = "+PONG\r\n"

const subscribedModePingResponse = "*2\r\n$4\r\npong\r\n$0\r\n\r\n"

func HandlePing(connection net.Conn, command *RedisCommand) string {
	if isConnectionInSubscribedMode(connection) {
		return subscribedModePingResponse
	}

	return normalPingResponse
}
