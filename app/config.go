package main

import (
	"flag"
	"strings"
)

// Config stores the configuration for the Redis server
type Config struct {
	Dir        string
	DbFilename string
	Port       int
	IsReplica  bool
	MasterHost string
	MasterPort string
}

var serverConfig Config

// ParseConfig parses the command-line arguments and sets the server configuration
func ParseConfig() {
	var replicaOf string
	flag.StringVar(&serverConfig.Dir, "dir", "", "the path to the directory where the RDB file is stored")
	flag.StringVar(&serverConfig.DbFilename, "dbfilename", "", "the name of the RDB file")
	flag.IntVar(&serverConfig.Port, "port", 6379, "the port number for the server to listen on")
	flag.StringVar(&replicaOf, "replicaof", "", "master host and port for replication (format: 'host port')")
	flag.Parse()

	// Parse replicaof flag if provided
	if replicaOf != "" {
		// Split on space to get host and port
		parts := strings.Split(replicaOf, " ")
		if len(parts) == 2 {
			serverConfig.IsReplica = true
			serverConfig.MasterHost = parts[0]
			serverConfig.MasterPort = parts[1]
		}
	}
}

// GetConfig returns the current server configuration
func GetConfig() Config {
	return serverConfig
}
