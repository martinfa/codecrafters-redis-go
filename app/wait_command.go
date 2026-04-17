package main

import (
	"fmt"
	"strconv"
	"time"
)

// HandleWait processes a WAIT command and returns a RESP integer response.
func HandleWait(command *RedisCommand) string {
	if len(command.Args) != 2 {
		return "-ERR wrong number of arguments for 'wait' command\r\n"
	}

	requiredReplicaCount, parseError := strconv.Atoi(command.Args[0])
	if parseError != nil {
		return fmt.Sprintf("-ERR invalid numreplicas value %q\r\n", command.Args[0])
	}

	timeoutMilliseconds, parseError := strconv.Atoi(command.Args[1])
	if parseError != nil {
		return fmt.Sprintf("-ERR invalid timeout value %q\r\n", command.Args[1])
	}

	targetReplicationOffset := CurrentMasterReplicationOffset()
	if targetReplicationOffset > 0 {
		requestReplicaAcknowledgements()
	}

	waitDeadline := time.Now().Add(time.Duration(timeoutMilliseconds) * time.Millisecond)
	acknowledgedReplicaCount := CountReplicasAcknowledgingOffset(targetReplicationOffset)
	for acknowledgedReplicaCount < requiredReplicaCount && time.Now().Before(waitDeadline) {
		time.Sleep(10 * time.Millisecond)
		acknowledgedReplicaCount = CountReplicasAcknowledgingOffset(targetReplicationOffset)
	}

	return fmt.Sprintf(":%d\r\n", acknowledgedReplicaCount)
}

func requestReplicaAcknowledgements() {
	getAcknowledgementCommand := []byte("*3\r\n$8\r\nREPLCONF\r\n$6\r\nGETACK\r\n$1\r\n*\r\n")
	for _, replicaConnection := range ReplicaConnectionsSnapshot() {
		if _, writeError := replicaConnection.Write(getAcknowledgementCommand); writeError != nil {
			fmt.Printf("Error requesting acknowledgement from replica %s: %v\n", replicaConnection.RemoteAddr(), writeError)
		}
	}
}
