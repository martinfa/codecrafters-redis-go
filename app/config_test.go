package main

import (
	"flag"
	"os"
	"testing"
)

func TestParseConfig_DefaultPort(t *testing.T) {
	// Save original command line arguments
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Set up test arguments with no port flag
	os.Args = []string{"test"}

	// Reset flag state
	flag.CommandLine = flag.NewFlagSet("test", flag.ExitOnError)

	// Reset config
	serverConfig = Config{}

	// Parse config
	ParseConfig()

	// Check that default port is set
	config := GetConfig()
	if config.Port != 6379 {
		t.Errorf("Expected default port 6379, got %d", config.Port)
	}
}

func TestParseConfig_CustomPort(t *testing.T) {
	// Save original command line arguments
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Set up test arguments with custom port
	os.Args = []string{"test", "-port", "6380"}

	// Reset flag state
	flag.CommandLine = flag.NewFlagSet("test", flag.ExitOnError)

	// Reset config
	serverConfig = Config{}

	// Parse config
	ParseConfig()

	// Check that custom port is set
	config := GetConfig()
	if config.Port != 6380 {
		t.Errorf("Expected port 6380, got %d", config.Port)
	}
}

func TestParseConfig_WithOtherFlags(t *testing.T) {
	// Save original command line arguments
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Set up test arguments with port and other flags
	os.Args = []string{"test", "-port", "12345", "-dir", "/tmp/redis", "-dbfilename", "test.rdb"}

	// Reset flag state
	flag.CommandLine = flag.NewFlagSet("test", flag.ExitOnError)

	// Reset config
	serverConfig = Config{}

	// Parse config
	ParseConfig()

	// Check that all values are set correctly
	config := GetConfig()
	if config.Port != 12345 {
		t.Errorf("Expected port 12345, got %d", config.Port)
	}
	if config.Dir != "/tmp/redis" {
		t.Errorf("Expected dir '/tmp/redis', got '%s'", config.Dir)
	}
	if config.DbFilename != "test.rdb" {
		t.Errorf("Expected dbfilename 'test.rdb', got '%s'", config.DbFilename)
	}
}

func TestParseConfig_ReplicaOf(t *testing.T) {
	// Save original command line arguments
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Set up test arguments with replicaof flag
	os.Args = []string{"test", "-replicaof", "localhost 6379"}

	// Reset flag state
	flag.CommandLine = flag.NewFlagSet("test", flag.ExitOnError)

	// Reset config
	serverConfig = Config{}

	// Parse config
	ParseConfig()

	// Check that replica config is set correctly
	config := GetConfig()
	if !config.IsReplica {
		t.Errorf("Expected IsReplica to be true, got false")
	}
	if config.MasterHost != "localhost" {
		t.Errorf("Expected MasterHost 'localhost', got '%s'", config.MasterHost)
	}
	if config.MasterPort != "6379" {
		t.Errorf("Expected MasterPort '6379', got '%s'", config.MasterPort)
	}
}

func TestParseConfig_NoReplicaOf(t *testing.T) {
	// Save original command line arguments
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Set up test arguments without replicaof flag
	os.Args = []string{"test", "-port", "6380"}

	// Reset flag state
	flag.CommandLine = flag.NewFlagSet("test", flag.ExitOnError)

	// Reset config
	serverConfig = Config{}

	// Parse config
	ParseConfig()

	// Check that replica config is not set
	config := GetConfig()
	if config.IsReplica {
		t.Errorf("Expected IsReplica to be false, got true")
	}
	if config.MasterHost != "" {
		t.Errorf("Expected MasterHost to be empty, got '%s'", config.MasterHost)
	}
	if config.MasterPort != "" {
		t.Errorf("Expected MasterPort to be empty, got '%s'", config.MasterPort)
	}
}
