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
