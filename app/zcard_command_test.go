package main

import "testing"

func TestParseCommandType_ZCARDReturnsKnownCommand(t *testing.T) {
	commandType := ParseCommandType("ZCARD")
	if commandType != CmdZCARD {
		t.Fatalf("expected ZCARD to parse as CmdZCARD, got %v", commandType)
	}
}

func TestParseZcardCommandArguments(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectedKey   string
		expectedError string
	}{
		{
			name:        "parses key",
			args:        []string{"zset_key"},
			expectedKey: "zset_key",
		},
		{
			name:          "rejects missing arguments",
			args:          []string{},
			expectedError: "-ERR wrong number of arguments for 'zcard' command\r\n",
		},
		{
			name:          "rejects too many arguments",
			args:          []string{"zset_key", "extra"},
			expectedError: "-ERR wrong number of arguments for 'zcard' command\r\n",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			key, errorResponse := parseZcardCommandArguments(&RedisCommand{
				Type: CmdZCARD,
				Args: testCase.args,
			})

			if errorResponse != testCase.expectedError {
				t.Errorf("parseZcardCommandArguments() error = %q, expected %q", errorResponse, testCase.expectedError)
			}

			if key != testCase.expectedKey {
				t.Errorf("parseZcardCommandArguments() key = %q, expected %q", key, testCase.expectedKey)
			}
		})
	}
}

func TestHandleZcardCodecraftersStageExample(t *testing.T) {
	resetSortedSetTestState(t)

	addSortedSetMembers(t, "zset_key",
		SortedSetMember{Member: "zset_member1", Score: 20.0},
		SortedSetMember{Member: "zset_member2", Score: 30.1},
		SortedSetMember{Member: "zset_member3", Score: 40.2},
		SortedSetMember{Member: "zset_member4", Score: 50.3},
	)

	result := HandleZcard(&RedisCommand{
		Type: CmdZCARD,
		Args: []string{"zset_key"},
	})
	if result != ":4\r\n" {
		t.Errorf("HandleZcard() = %q, expected %q", result, ":4\r\n")
	}

	updateResult := HandleZadd(&RedisCommand{
		Type: CmdZADD,
		Args: []string{"zset_key", "100.0", "zset_member1"},
	})
	if updateResult != ":0\r\n" {
		t.Errorf("HandleZadd() = %q, expected %q", updateResult, ":0\r\n")
	}

	resultAfterUpdate := HandleZcard(&RedisCommand{
		Type: CmdZCARD,
		Args: []string{"zset_key"},
	})
	if resultAfterUpdate != ":4\r\n" {
		t.Errorf("HandleZcard() after score update = %q, expected %q", resultAfterUpdate, ":4\r\n")
	}
}

func TestHandleZcardMissingKeyReturnsZero(t *testing.T) {
	resetSortedSetTestState(t)

	result := HandleZcard(&RedisCommand{
		Type: CmdZCARD,
		Args: []string{"missing_key"},
	})
	if result != ":0\r\n" {
		t.Errorf("HandleZcard() = %q, expected %q", result, ":0\r\n")
	}
}

func TestHandleZcardArgumentParsing(t *testing.T) {
	resetSortedSetTestState(t)

	result := HandleZcard(&RedisCommand{
		Type: CmdZCARD,
		Args: []string{},
	})
	if result != "-ERR wrong number of arguments for 'zcard' command\r\n" {
		t.Errorf("HandleZcard() = %q, expected %q", result, "-ERR wrong number of arguments for 'zcard' command\r\n")
	}
}
