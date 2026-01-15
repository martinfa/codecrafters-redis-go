package main

import (
	"flag"
	"strings"
)

// Config stores the configuration for the Redis server
type Config struct {
	Dir              string
	DbFilename       string
	Port             int
	IsReplica        bool
	MasterHost       string
	MasterPort       string
	MasterReplId     string
	MasterReplOffset int
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

	// Initialize replication values (hardcoded for this stage)
	serverConfig.MasterReplId = "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb"
	serverConfig.MasterReplOffset = 0

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
