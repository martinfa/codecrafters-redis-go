package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
)

// Ensures gofmt doesn't remove the "net" and "os" imports in stage 1 (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit

func listen(conn net.Conn, channel chan []byte) {
	defer conn.Close()   // Ensure the connection for this client is closed when listen exits
	defer close(channel) // Close the channel to signal eventReactor to stop

	buffer := make([]byte, 1024)
	for {
		// Try to read data from the connection.
		n, err := conn.Read(buffer)
		if err != nil {
			if err == io.EOF {
				// Connection closed by client, normal termination for this loop
				fmt.Printf("Client %s closed the connection.\n", conn.RemoteAddr())
			} else {
				// Some other read error occurred
				fmt.Printf("Error reading from connection %s: %s\n", conn.RemoteAddr(), err.Error())
			}
			break // Exit the loop
		}

		// Send only the part of the buffer that contains data
		dataRead := make([]byte, n) // Create a new slice with the exact size
		copy(dataRead, buffer[:n])  // Copy the data
		channel <- dataRead
	}
}

func ping_command(conn net.Conn) {
	_, err := conn.Write([]byte("+PONG\r\n"))
	if err != nil {
		fmt.Printf("Error writing PONG to connection %s: %s\n", conn.RemoteAddr(), err.Error())
	}
}

func eventReactor(channel chan []byte, conn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		// Wait for data from the listen goroutine
		// The ok variable will be false if the channel is closed
		buffer, ok := <-channel
		if !ok {
			fmt.Printf("Channel closed for %s, eventReactor exiting.\n", conn.RemoteAddr())
			break // Exit loop if channel is closed
		}

		// Convert buffer to string and trim whitespace
		command := strings.TrimSpace(string(buffer))
		// Log the raw command received for debugging, including for empty strings after trim
		fmt.Printf("Received from %s (trimmed): \"%s\"\n", conn.RemoteAddr(), command)

		// For this stage, we respond PONG to any command chunk received that makes it here.
		// The problem description implies multiple PINGs on the same connection,
		// and hardcoding PONG. RESP parsing is for later.
		// If command is empty after trim (e.g. just \r\n), we might still PONG or not.
		// The current test passes if any data from client 1 gets a PONG.
		// To be safe and align with "any data gets a PONG for this stage":
		ping_command(conn)
	}
}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage

	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379:", err.Error())
		os.Exit(1)
	}
	defer l.Close()
	fmt.Println("Server listening on 0.0.0.0:6379")

	for { // Loop indefinitely to accept multiple connections
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err.Error())
			// If the listener is closed, Accept will return an error, and the loop will break.
			// For other temporary errors, we might want to continue.
			// Checking for net.ErrClosed specifically can make this more robust.
			if _, ok := err.(*net.OpError); ok && err.Error() == "use of closed network connection" {
				fmt.Println("Listener closed, server shutting down.")
				break
			}
			continue
		}

		fmt.Printf("Accepted new connection from %s\n", conn.RemoteAddr())

		// For each connection, set up its own channel and WaitGroup
		clientChannel := make(chan []byte)
		var clientWg sync.WaitGroup // Each client connection gets its own WaitGroup

		clientWg.Add(1) // We expect one eventReactor goroutine for this client
		go listen(conn, clientChannel)
		// Pass the per-client WaitGroup to its eventReactor
		go eventReactor(clientChannel, conn, &clientWg)

		// The main goroutine does NOT wait here using clientWg.Wait().
		// It loops back to accept the next client.
		// clientWg.Wait() would be used if we wanted to do something *after* a specific client is done,
		// but not to block accepting other clients.
	}
}
