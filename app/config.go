package main

import (
	"flag"
)

// Config stores the configuration for the Redis server
type Config struct {
	Dir        string
	DbFilename string
	Port       int
}

var serverConfig Config

// ParseConfig parses the command-line arguments and sets the server configuration
func ParseConfig() {
	flag.StringVar(&serverConfig.Dir, "dir", "", "the path to the directory where the RDB file is stored")
	flag.StringVar(&serverConfig.DbFilename, "dbfilename", "", "the name of the RDB file")
	flag.IntVar(&serverConfig.Port, "port", 6379, "the port number for the server to listen on")
	flag.Parse()
}

// GetConfig returns the current server configuration
func GetConfig() Config {
	return serverConfig
}
