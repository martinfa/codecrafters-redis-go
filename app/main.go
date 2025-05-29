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

	buffer := make([]byte, 1024)
	for {
		// Try to read data from the connection.
		n, err := conn.Read(buffer)
		if err != nil {
			if err == io.EOF {
				// Connection closed by client, normal termination for this loop
				fmt.Println("Client closed the connection.")
				close(channel) // Close the channel to signal eventReactor to stop
			} else {
				// Some other read error occurred
				fmt.Println("Error reading from connection:", err.Error())
				close(channel) // Also close on other errors
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
		fmt.Println("Error writing PONG to connection:", err.Error())
	}
}

func eventReactor(channel chan []byte, conn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		// Wait for data from the listen goroutine
		// The ok variable will be false if the channel is closed
		buffer, ok := <-channel
		if !ok {
			fmt.Println("Channel closed, eventReactor exiting.")
			break // Exit loop if channel is closed
		}

		// Convert buffer to string and trim whitespace
		command := strings.TrimSpace(string(buffer))
		fmt.Println("Received command:", command) // For debugging

		// For this stage, we respond PONG to any command (including empty strings after trim)
		// In later stages, you'll parse specific commands like PING.
		if command == "PING" { // Specific check for PING
			ping_command(conn)
		} else {
			// For now, still respond PONG to anything else, or handle appropriately
			// As per instructions for this stage, any input should get a PONG.
			// If the input after TrimSpace is empty (e.g. client sent only \r\n),
			// we might still want to PONG or decide how to handle it.
			// For simplicity, let's PONG if anything (even empty string) was received and channel not closed.
			ping_command(conn)
		}
	}
}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage

	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	defer l.Close()

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}
	defer conn.Close()

	channel := make(chan []byte)
	var wg sync.WaitGroup

	wg.Add(1)
	go listen(conn, channel)
	go eventReactor(channel, conn, &wg)

	wg.Wait()

	fmt.Println("Server shutting down gracefully.")
}
