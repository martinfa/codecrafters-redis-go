package main

import (
	"net"
	"sync"
)

const errExecWithoutMulti = "-ERR EXEC without MULTI\r\n"
const errExecWithoutQueuedCommands = "*0\r\n"

type connectionTransactionState struct {
	inTransaction  bool
	queuedCommands []*RedisCommand
}

var (
	connectionTransactionMutex  sync.Mutex
	connectionTransactionStates = make(map[net.Conn]*connectionTransactionState)
)

func ResetConnectionTransactionStatesForTest() {
	connectionTransactionMutex.Lock()
	defer connectionTransactionMutex.Unlock()

	connectionTransactionStates = make(map[net.Conn]*connectionTransactionState)
}

func getConnectionTransactionState(connection net.Conn) *connectionTransactionState {
	connectionTransactionMutex.Lock()
	defer connectionTransactionMutex.Unlock()

	state, exists := connectionTransactionStates[connection]
	if !exists {
		state = &connectionTransactionState{}
		connectionTransactionStates[connection] = state
	}

	return state
}

func RemoveConnectionTransactionState(connection net.Conn) {
	connectionTransactionMutex.Lock()
	defer connectionTransactionMutex.Unlock()

	delete(connectionTransactionStates, connection)
}

func parseMultiCommandArguments(command *RedisCommand) (errorResponse string) {
	if len(command.Args) != 0 {
		return "-ERR wrong number of arguments for 'multi' command\r\n"
	}

	return ""
}

func HandleMulti(connection net.Conn, command *RedisCommand) string {
	errorResponse := parseMultiCommandArguments(command)
	if errorResponse != "" {
		return errorResponse
	}

	transactionState := getConnectionTransactionState(connection)
	transactionState.inTransaction = true

	return "+OK\r\n"
}

func parseExecCommandArguments(command *RedisCommand) (errorResponse string) {
	if len(command.Args) != 0 {
		return "-ERR wrong number of arguments for 'exec' command\r\n"
	}

	return ""
}

func HandleExec(connection net.Conn, command *RedisCommand) string {
	errorResponse := parseExecCommandArguments(command)
	if errorResponse != "" {
		return errorResponse
	}

	transactionState := getConnectionTransactionState(connection)
	if !transactionState.inTransaction {
		return errExecWithoutMulti
	}

	if len(transactionState.queuedCommands) == 0 {
		RemoveConnectionTransactionState(connection)
		return errExecWithoutQueuedCommands
	}

	return "*0\r\n"
}
