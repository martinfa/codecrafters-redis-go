package main

import (
	"testing"
	"time"
)

func resetBlpopTestState(t *testing.T) {
	t.Helper()
	GetInstance().cache = make(map[string]CacheItem)
	SetBlockingBlpopRegistryForTest(NewBlockingBlpopRegistry())
}

func formatExpectedBlpopResponse(listKey string, element string) string {
	return encodeBlpopResponse(listKey, element)
}

func TestParseBlpopCommandArguments(t *testing.T) {
	tests := []struct {
		name                     string
		args                     []string
		expectedListKey          string
		expectedTimeoutSeconds   string
		expectedError            string
	}{
		{
			name:                   "parses list key and timeout",
			args:                   []string{"list_key", "0"},
			expectedListKey:        "list_key",
			expectedTimeoutSeconds: "0",
		},
		{
			name:                   "parses fractional timeout",
			args:                   []string{"list_key", "0.5"},
			expectedListKey:        "list_key",
			expectedTimeoutSeconds: "0.5",
		},
		{
			name:          "rejects no arguments",
			args:          []string{},
			expectedError: "-ERR wrong number of arguments for 'blpop' command\r\n",
		},
		{
			name:          "rejects missing timeout",
			args:          []string{"list_key"},
			expectedError: "-ERR wrong number of arguments for 'blpop' command\r\n",
		},
		{
			name:          "rejects too many arguments",
			args:          []string{"list_key", "0", "extra"},
			expectedError: "-ERR wrong number of arguments for 'blpop' command\r\n",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			listKey, timeoutSeconds, errorResponse := parseBlpopCommandArguments(&RedisCommand{
				Type: CmdBLPOP,
				Args: testCase.args,
			})

			if errorResponse != testCase.expectedError {
				t.Errorf("parseBlpopCommandArguments() error = %q, expected %q", errorResponse, testCase.expectedError)
			}

			if listKey != testCase.expectedListKey {
				t.Errorf("parseBlpopCommandArguments() listKey = %q, expected %q", listKey, testCase.expectedListKey)
			}

			if timeoutSeconds != testCase.expectedTimeoutSeconds {
				t.Errorf("parseBlpopCommandArguments() timeoutSeconds = %q, expected %q", timeoutSeconds, testCase.expectedTimeoutSeconds)
			}
		})
	}
}

func TestHandleBlpopReturnsImmediatelyWhenListHasElements(t *testing.T) {
	resetBlpopTestState(t)

	HandleRpush(&RedisCommand{
		Type: CmdRPUSH,
		Args: []string{"list_key", "foo"},
	})

	result := HandleBlpop(&RedisCommand{
		Type: CmdBLPOP,
		Args: []string{"list_key", "0"},
	})

	expected := formatExpectedBlpopResponse("list_key", "foo")
	if result != expected {
		t.Errorf("HandleBlpop() = %q, expected %q", result, expected)
	}
}

func TestHandleBlpopCodecraftersStageExample(t *testing.T) {
	resetBlpopTestState(t)

	resultChannel := make(chan string, 1)
	go func() {
		resultChannel <- HandleBlpop(&RedisCommand{
			Type: CmdBLPOP,
			Args: []string{"list_key", "0"},
		})
	}()

	time.Sleep(50 * time.Millisecond)

	rpushResult := HandleRpush(&RedisCommand{
		Type: CmdRPUSH,
		Args: []string{"list_key", "foo"},
	})
	if rpushResult != ":1\r\n" {
		t.Fatalf("HandleRpush() = %q, expected %q", rpushResult, ":1\r\n")
	}

	select {
	case result := <-resultChannel:
		expected := formatExpectedBlpopResponse("list_key", "foo")
		if result != expected {
			t.Errorf("HandleBlpop() = %q, expected %q", result, expected)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for blocking BLPOP response")
	}
}

func TestHandleBlpopRespondsToLongestWaitingClientFirst(t *testing.T) {
	resetBlpopTestState(t)

	firstClientResultChannel := make(chan string, 1)
	secondClientResultChannel := make(chan string, 1)

	go func() {
		firstClientResultChannel <- HandleBlpop(&RedisCommand{
			Type: CmdBLPOP,
			Args: []string{"another_list_key", "0"},
		})
	}()

	time.Sleep(20 * time.Millisecond)

	go func() {
		secondClientResultChannel <- HandleBlpop(&RedisCommand{
			Type: CmdBLPOP,
			Args: []string{"another_list_key", "0"},
		})
	}()

	time.Sleep(50 * time.Millisecond)

	HandleRpush(&RedisCommand{
		Type: CmdRPUSH,
		Args: []string{"another_list_key", "foo"},
	})

	select {
	case firstClientResult := <-firstClientResultChannel:
		expected := formatExpectedBlpopResponse("another_list_key", "foo")
		if firstClientResult != expected {
			t.Errorf("first HandleBlpop() = %q, expected %q", firstClientResult, expected)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for first blocking BLPOP response")
	}

	select {
	case <-secondClientResultChannel:
		t.Fatal("second client should still be blocked after one element is popped")
	case <-time.After(100 * time.Millisecond):
	}
}

func TestHandleBlpopRemovesPoppedElementFromList(t *testing.T) {
	resetBlpopTestState(t)

	HandleRpush(&RedisCommand{
		Type: CmdRPUSH,
		Args: []string{"list_key", "foo", "bar"},
	})

	result := HandleBlpop(&RedisCommand{
		Type: CmdBLPOP,
		Args: []string{"list_key", "0"},
	})
	if result != formatExpectedBlpopResponse("list_key", "foo") {
		t.Fatalf("HandleBlpop() = %q, expected %q", result, formatExpectedBlpopResponse("list_key", "foo"))
	}

	rangeResult := HandleLrange(&RedisCommand{
		Type: CmdLRANGE,
		Args: []string{"list_key", "0", "-1"},
	})
	if rangeResult != formatExpectedLrangeResponse("bar") {
		t.Errorf("HandleLrange() = %q, expected %q", rangeResult, formatExpectedLrangeResponse("bar"))
	}
}

func TestHandleBlpopArgumentParsing(t *testing.T) {
	resetBlpopTestState(t)

	result := HandleBlpop(&RedisCommand{
		Type: CmdBLPOP,
		Args: []string{"list_key"},
	})

	if result != "-ERR wrong number of arguments for 'blpop' command\r\n" {
		t.Errorf("HandleBlpop() = %q, expected %q", result, "-ERR wrong number of arguments for 'blpop' command\r\n")
	}
}

func TestHandleBlpopRejectsInvalidTimeout(t *testing.T) {
	resetBlpopTestState(t)

	result := HandleBlpop(&RedisCommand{
		Type: CmdBLPOP,
		Args: []string{"list_key", "not-a-number"},
	})

	if result != "-ERR value is not an integer or out of range\r\n" {
		t.Errorf("HandleBlpop() = %q, expected %q", result, "-ERR value is not an integer or out of range\r\n")
	}
}

func TestHandleBlpopReturnsNullArrayOnTimeout(t *testing.T) {
	resetBlpopTestState(t)

	startTime := time.Now()
	result := HandleBlpop(&RedisCommand{
		Type: CmdBLPOP,
		Args: []string{"list_key", "0.1"},
	})
	elapsed := time.Since(startTime)

	if result != blockingBlpopTimeoutResponse {
		t.Errorf("HandleBlpop() = %q, expected %q", result, blockingBlpopTimeoutResponse)
	}

	if elapsed < 90*time.Millisecond {
		t.Errorf("expected to block for at least 90ms, got %v", elapsed)
	}
}

func TestHandleBlpopWaitsForPushBeforeTimeout(t *testing.T) {
	resetBlpopTestState(t)

	resultChannel := make(chan string, 1)
	go func() {
		resultChannel <- HandleBlpop(&RedisCommand{
			Type: CmdBLPOP,
			Args: []string{"list_key", "1"},
		})
	}()

	time.Sleep(50 * time.Millisecond)

	HandleRpush(&RedisCommand{
		Type: CmdRPUSH,
		Args: []string{"list_key", "foo"},
	})

	select {
	case result := <-resultChannel:
		expected := formatExpectedBlpopResponse("list_key", "foo")
		if result != expected {
			t.Errorf("HandleBlpop() = %q, expected %q", result, expected)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for BLPOP response before timeout")
	}
}

func TestHandleBlpopUnblocksTimedOutWaiterDoesNotReceiveLaterPush(t *testing.T) {
	resetBlpopTestState(t)

	resultChannel := make(chan string, 1)
	go func() {
		resultChannel <- HandleBlpop(&RedisCommand{
			Type: CmdBLPOP,
			Args: []string{"list_key", "0.1"},
		})
	}()

	select {
	case result := <-resultChannel:
		if result != blockingBlpopTimeoutResponse {
			t.Fatalf("HandleBlpop() = %q, expected %q", result, blockingBlpopTimeoutResponse)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for BLPOP timeout response")
	}

	HandleRpush(&RedisCommand{
		Type: CmdRPUSH,
		Args: []string{"list_key", "foo"},
	})

	rangeResult := HandleLrange(&RedisCommand{
		Type: CmdLRANGE,
		Args: []string{"list_key", "0", "-1"},
	})
	if rangeResult != formatExpectedLrangeResponse("foo") {
		t.Errorf("HandleLrange() = %q, expected element to remain on list after timed out BLPOP", rangeResult)
	}
}
