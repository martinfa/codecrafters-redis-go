package main

import (
	"fmt"
	"net"
	"strings"
	"sync"
)

const errExecWithoutMulti = "-ERR EXEC without MULTI\r\n"
const errDiscardWithoutMulti = "-ERR DISCARD without MULTI\r\n"
const errExecWithoutQueuedCommands = "*0\r\n"
const queuedCommandResponse = "+QUEUED\r\n"

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

func ShouldQueueCommandDuringTransaction(connection net.Conn, command *RedisCommand) bool {
	if command.Type == CmdMULTI || command.Type == CmdEXEC || command.Type == CmdDISCARD {
		return false
	}

	transactionState := getConnectionTransactionState(connection)
	return transactionState.inTransaction
}

func queueTransactionCommand(transactionState *connectionTransactionState, command *RedisCommand) string {
	transactionState.queuedCommands = append(transactionState.queuedCommands, &RedisCommand{
		Type: command.Type,
		Args: append([]string(nil), command.Args...),
	})

	return queuedCommandResponse
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

	responses := make([]string, 0, len(transactionState.queuedCommands))
	for _, queuedCommand := range transactionState.queuedCommands {
		responses = append(responses, executeConnectionCommand(connection, queuedCommand))
	}

	RemoveConnectionTransactionState(connection)

	return encodeExecResponse(responses)
}

func parseDiscardCommandArguments(command *RedisCommand) (errorResponse string) {
	if len(command.Args) != 0 {
		return "-ERR wrong number of arguments for 'discard' command\r\n"
	}

	return ""
}

func HandleDiscard(connection net.Conn, command *RedisCommand) string {
	errorResponse := parseDiscardCommandArguments(command)
	if errorResponse != "" {
		return errorResponse
	}

	transactionState := getConnectionTransactionState(connection)
	if !transactionState.inTransaction {
		return errDiscardWithoutMulti
	}

	RemoveConnectionTransactionState(connection)

	return "+OK\r\n"
}

func encodeExecResponse(responses []string) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("*%d\r\n", len(responses)))
	for _, response := range responses {
		builder.WriteString(response)
	}

	return builder.String()
}

func executeConnectionCommand(connection net.Conn, command *RedisCommand) string {
	switch command.Type {
	case CmdECHO:
		return HandleEcho(command)
	case CmdSET:
		return HandleSet(command)
	case CmdGET:
		return HandleGet(command)
	case CmdCONFIG:
		return HandleConfig(command)
	case CmdKEYS:
		return HandleKeys(command)
	case CmdINFO:
		return HandleInfo(command)
	case CmdPING:
		return "+PONG\r\n"
	case CmdREPLCONF:
		return "+OK\r\n"
	case CmdPSYNC:
		return HandlePsync(command, connection)
	case CmdWAIT:
		return HandleWait(command)
	case CmdTYPE:
		return HandleType(command)
	case CmdXADD:
		return HandleXadd(command)
	case CmdXRANGE:
		return HandleXrange(command)
	case CmdXREAD:
		return HandleXread(command)
	case CmdINCR:
		return HandleIncr(command)
	case CmdRPUSH:
		return HandleRpush(command)
	case CmdMULTI:
		return HandleMulti(connection, command)
	case CmdEXEC:
		return HandleExec(connection, command)
	default:
		return "-ERR unknown command\r\n"
	}
}

func HandleConnectionCommand(connection net.Conn, command *RedisCommand) string {
	if command.Type == CmdMULTI {
		return HandleMulti(connection, command)
	}

	if command.Type == CmdEXEC {
		return HandleExec(connection, command)
	}

	if command.Type == CmdDISCARD {
		return HandleDiscard(connection, command)
	}

	transactionState := getConnectionTransactionState(connection)
	if transactionState.inTransaction {
		return queueTransactionCommand(transactionState, command)
	}

	return executeConnectionCommand(connection, command)
}
