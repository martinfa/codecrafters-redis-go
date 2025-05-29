package main

import (
	"fmt"
	"io"
	"net"
	"os"
)

// Ensures gofmt doesn't remove the "net" and "os" imports in stage 1 (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit

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

	// Buffer to read client input, 1KB should be plenty for PINGs
	buffer := make([]byte, 1024)

	for {
		// Try to read data from the connection.
		// For this stage, we don't need to inspect the content of the buffer.
		// The act of receiving data is enough to trigger a PONG.
		_, err := conn.Read(buffer)
		if err != nil {
			if err == io.EOF {
				// Connection closed by client, normal termination for this loop
				fmt.Println("Client closed the connection.")
			} else {
				// Some other read error occurred
				fmt.Println("Error reading from connection:", err.Error())
			}
			break // Exit the loop
		}

		// Send back PONG
		_, err = conn.Write([]byte("+PONG\r\n"))
		if err != nil {
			fmt.Println("Error writing PONG to connection:", err.Error())
			break // Exit the loop if we can't write
		}
	}
}
