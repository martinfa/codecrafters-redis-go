package main

import (
	"testing"
	"time"
)

func resetXreadTestState(t *testing.T) {
	t.Helper()
	GetInstance().cache = make(map[string]CacheItem)
	SetEventBusForTest(NewEventBus())
}

func TestParseBlockingXreadCommand(t *testing.T) {
	streamKeys, startIDs, blockMilliseconds, errorResponse := parseXreadCommand(&RedisCommand{
		Type: CmdXREAD,
		Args: []string{"BLOCK", "1000", "streams", "stream_key", "0-1"},
	})

	if errorResponse != "" {
		t.Fatalf("parseXreadCommand() error = %q", errorResponse)
	}

	if blockMilliseconds != 1000 {
		t.Fatalf("blockMilliseconds = %d, expected 1000", blockMilliseconds)
	}

	if len(streamKeys) != 1 || streamKeys[0] != "stream_key" {
		t.Fatalf("streamKeys = %v, expected [stream_key]", streamKeys)
	}

	if len(startIDs) != 1 || startIDs[0] != "0-1" {
		t.Fatalf("startIDs = %v, expected [0-1]", startIDs)
	}
}

func TestHandleBlockingXreadReturnsImmediatelyWhenEntriesExist(t *testing.T) {
	resetXreadTestState(t)

	HandleXadd(&RedisCommand{
		Type: CmdXADD,
		Args: []string{"stream_key", "0-1", "temperature", "96"},
	})

	result := HandleXread(&RedisCommand{
		Type: CmdXREAD,
		Args: []string{"BLOCK", "1000", "STREAMS", "stream_key", "0-0"},
	})

	expected := formatExpectedXreadResponse(
		formatExpectedXreadStreamResponse(
			"stream_key",
			formatExpectedXreadEntryResponse("0-1", "temperature", "96"),
		),
	)
	if result != expected {
		t.Errorf("HandleXread() = %q, expected %q", result, expected)
	}
}

func TestHandleBlockingXreadWaitsForStreamChange(t *testing.T) {
	resetXreadTestState(t)

	HandleXadd(&RedisCommand{
		Type: CmdXADD,
		Args: []string{"stream_key", "0-1", "temperature", "96"},
	})

	resultChannel := make(chan string, 1)
	go func() {
		resultChannel <- HandleXread(&RedisCommand{
			Type: CmdXREAD,
			Args: []string{"BLOCK", "1000", "STREAMS", "stream_key", "0-1"},
		})
	}()

	time.Sleep(50 * time.Millisecond)

	HandleXadd(&RedisCommand{
		Type: CmdXADD,
		Args: []string{"stream_key", "0-2", "temperature", "95"},
	})

	select {
	case result := <-resultChannel:
		expected := formatExpectedXreadResponse(
			formatExpectedXreadStreamResponse(
				"stream_key",
				formatExpectedXreadEntryResponse("0-2", "temperature", "95"),
			),
		)
		if result != expected {
			t.Errorf("HandleXread() = %q, expected %q", result, expected)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for blocking XREAD response")
	}
}

func TestHandleBlockingXreadReturnsNullArrayOnTimeout(t *testing.T) {
	resetXreadTestState(t)

	HandleXadd(&RedisCommand{
		Type: CmdXADD,
		Args: []string{"stream_key", "0-1", "temperature", "96"},
	})

	startTime := time.Now()
	result := HandleXread(&RedisCommand{
		Type: CmdXREAD,
		Args: []string{"block", "100", "streams", "stream_key", "0-1"},
	})
	elapsed := time.Since(startTime)

	if result != blockingXreadTimeoutResponse {
		t.Errorf("HandleXread() = %q, expected %q", result, blockingXreadTimeoutResponse)
	}

	if elapsed < 90*time.Millisecond {
		t.Errorf("expected to block for at least 90ms, got %v", elapsed)
	}
}
