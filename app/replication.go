package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	replicas      []net.Conn
	replicasMutex sync.Mutex
)

// RegisterReplica adds a new replica connection to the list of replicas
func RegisterReplica(conn net.Conn) {
	if conn == nil {
		return
	}
	replicasMutex.Lock()
	defer replicasMutex.Unlock()
	replicas = append(replicas, conn)
	fmt.Printf("Registered new replica: %s. Total replicas: %d\n", conn.RemoteAddr(), len(replicas))
}

// PropagateCommand sends a command to all registered replicas
func PropagateCommand(command []byte) {
	replicasMutex.Lock()
	defer replicasMutex.Unlock()

	for _, replica := range replicas {
		_, err := replica.Write(command)
		if err != nil {
			fmt.Printf("Error propagating command to replica %s: %v\n", replica.RemoteAddr(), err)
			// In a real implementation, we might want to remove the failed replica
		} else {
			fmt.Printf("Propagated command to replica %s: %q\n", replica.RemoteAddr(), string(command))
		}
	}
}

// InitiateHandshake handles the replication handshake with the master server
func InitiateHandshake(config Config) error {
	if !config.IsReplica {
		return nil
	}

	masterAddress := net.JoinHostPort(config.MasterHost, config.MasterPort)
	fmt.Printf("Connecting to master at %s...\n", masterAddress)

	conn, err := net.DialTimeout("tcp", masterAddress, 5*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to master: %w", err)
	}

	reader := bufio.NewReader(conn)

	// Step 1: Send PING
	if err := sendCommand(conn, reader, "*1\r\n$4\r\nPING\r\n"); err != nil {
		return fmt.Errorf("PING failed: %w", err)
	}
	fmt.Println("Handshake: Received PONG")

	// Step 2: Send REPLCONF listening-port <PORT>
	portStr := fmt.Sprintf("%d", config.Port)
	replConfPort := fmt.Sprintf("*3\r\n$8\r\nREPLCONF\r\n$14\r\nlistening-port\r\n$%d\r\n%s\r\n", len(portStr), portStr)
	if err := sendCommand(conn, reader, replConfPort); err != nil {
		return fmt.Errorf("REPLCONF listening-port failed: %w", err)
	}
	fmt.Println("Handshake: Received OK for listening-port")

	// Step 3: Send REPLCONF capa psync2
	replConfCapa := "*3\r\n$8\r\nREPLCONF\r\n$4\r\ncapa\r\n$6\r\npsync2\r\n"
	if err := sendCommand(conn, reader, replConfCapa); err != nil {
		return fmt.Errorf("REPLCONF capa failed: %w", err)
	}
	fmt.Println("Handshake: Received OK for capa")

	// Step 4: Send PSYNC ? -1
	psyncCmd := "*3\r\n$5\r\nPSYNC\r\n$1\r\n?\r\n$2\r\n-1\r\n"
	if err := sendCommand(conn, reader, psyncCmd); err != nil {
		return fmt.Errorf("PSYNC failed: %w", err)
	}
	fmt.Println("Handshake: Received FULLRESYNC (ignored for now)")

	if err := readReplicationHandshakeRDBSnapshot(reader); err != nil {
		return fmt.Errorf("read RDB snapshot from master: %w", err)
	}

	// Start processing propagated commands from master. Reads must use the same
	// bufio.Reader as the handshake: otherwise RDB bytes buffered after +FULLRESYNC
	// are invisible to conn.Read and the replication stream desynchronizes.
	masterChannel := make(chan []byte)
	var masterWg sync.WaitGroup
	masterWg.Add(1)

	go listenMasterReplicationConnection(conn, reader, masterChannel)
	processedReplicationCommandBytes := 0
	go eventReactor(masterChannel, conn, &masterWg, true, &processedReplicationCommandBytes)

	return nil
}

// FormatReplicaReplconfAcknowledgementRESPArray builds *3 REPLCONF ACK <offset> for the replication stream.
func FormatReplicaReplconfAcknowledgementRESPArray(processedReplicationCommandByteOffset int) string {
	offsetDigits := strconv.Itoa(processedReplicationCommandByteOffset)
	return fmt.Sprintf("*3\r\n$8\r\nREPLCONF\r\n$3\r\nACK\r\n$%d\r\n%s\r\n", len(offsetDigits), offsetDigits)
}

// readReplicationHandshakeRDBSnapshot consumes $<n>\r\n followed by n bytes of RDB payload
// that the master sends immediately after +FULLRESYNC.
func readReplicationHandshakeRDBSnapshot(reader *bufio.Reader) error {
	firstByte, err := reader.ReadByte()
	if err != nil {
		return fmt.Errorf("read before RDB bulk: %w", err)
	}
	if firstByte != '$' {
		return fmt.Errorf("expected RDB bulk to start with '$', got %q", firstByte)
	}
	lengthLine, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("read RDB bulk length line: %w", err)
	}
	lengthLine = strings.TrimRight(lengthLine, "\r\n")
	byteCount, err := strconv.Atoi(lengthLine)
	if err != nil {
		return fmt.Errorf("parse RDB bulk length %q: %w", lengthLine, err)
	}
	if _, err := io.CopyN(io.Discard, reader, int64(byteCount)); err != nil {
		return fmt.Errorf("discard RDB bulk payload: %w", err)
	}
	return nil
}

func listenMasterReplicationConnection(connection net.Conn, bufferedReader *bufio.Reader, channel chan []byte) {
	defer connection.Close()
	defer close(channel)

	readBuffer := make([]byte, 1024)
	for {
		bytesRead, readError := bufferedReader.Read(readBuffer)
		if readError != nil {
			if readError == io.EOF {
				fmt.Printf("Client %s closed the connection.\n", connection.RemoteAddr())
			} else {
				fmt.Printf("Error reading from connection %s: %s\n", connection.RemoteAddr(), readError.Error())
			}
			break
		}
		dataRead := make([]byte, bytesRead)
		copy(dataRead, readBuffer[:bytesRead])
		channel <- dataRead
	}
}

func sendCommand(conn net.Conn, reader *bufio.Reader, command string) error {
	_, err := fmt.Fprint(conn, command)
	if err != nil {
		return err
	}

	// Wait for response (+PONG or +OK)
	response, err := reader.ReadString('\n')
	if err != nil {
		return err
	}

	// For simple strings, we just check if it's not an error (-)
	if response[0] == '-' {
		return fmt.Errorf("master returned error: %s", response)
	}

	return nil
}
