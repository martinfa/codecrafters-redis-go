package main

import (
	"net"
	"sync"
)

var (
	connectionWriteMutexes sync.Mutex
	connectionWriteLocks   = make(map[net.Conn]*sync.Mutex)
)

func ResetConnectionWriteMutexesForTest() {
	connectionWriteMutexes.Lock()
	defer connectionWriteMutexes.Unlock()

	connectionWriteLocks = make(map[net.Conn]*sync.Mutex)
}

func getConnectionWriteMutex(connection net.Conn) *sync.Mutex {
	connectionWriteMutexes.Lock()
	defer connectionWriteMutexes.Unlock()

	mutex, exists := connectionWriteLocks[connection]
	if !exists {
		mutex = &sync.Mutex{}
		connectionWriteLocks[connection] = mutex
	}

	return mutex
}

func RemoveConnectionWriteMutex(connection net.Conn) {
	connectionWriteMutexes.Lock()
	defer connectionWriteMutexes.Unlock()

	delete(connectionWriteLocks, connection)
}

func WriteToConnection(connection net.Conn, response string) error {
	mutex := getConnectionWriteMutex(connection)
	mutex.Lock()
	defer mutex.Unlock()

	_, writeError := connection.Write([]byte(response))

	return writeError
}
