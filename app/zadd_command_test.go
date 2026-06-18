package main

import (
	"testing"
)

func resetZaddTestState(t *testing.T) {
	t.Helper()
	GetInstance().cache = make(map[string]CacheItem)
}

func TestParseCommandType_ZADDReturnsKnownCommand(t *testing.T) {
	commandType := ParseCommandType("ZADD")
	if commandType != CmdZADD {
		t.Fatalf("expected ZADD to parse as CmdZADD, got %v", commandType)
	}
}

func TestParseZaddCommandArguments(t *testing.T) {
	tests := []struct {
		name            string
		args            []string
		expectedKey     string
		expectedScore   float64
		expectedMember  string
		expectedError   string
	}{
		{
			name:           "parses key score and member",
			args:           []string{"zset_key", "10.0", "zset_member"},
			expectedKey:    "zset_key",
			expectedScore:  10.0,
			expectedMember: "zset_member",
		},
		{
			name:          "parses integer score as float",
			args:          []string{"zset_key", "10", "zset_member"},
			expectedKey:   "zset_key",
			expectedScore: 10.0,
			expectedMember: "zset_member",
		},
		{
			name:          "rejects missing arguments",
			args:          []string{},
			expectedError: "-ERR wrong number of arguments for 'zadd' command\r\n",
		},
		{
			name:          "rejects key only",
			args:          []string{"zset_key"},
			expectedError: "-ERR wrong number of arguments for 'zadd' command\r\n",
		},
		{
			name:          "rejects key and score only",
			args:          []string{"zset_key", "10.0"},
			expectedError: "-ERR wrong number of arguments for 'zadd' command\r\n",
		},
		{
			name:          "rejects too many arguments for this stage",
			args:          []string{"zset_key", "10.0", "zset_member", "extra"},
			expectedError: "-ERR wrong number of arguments for 'zadd' command\r\n",
		},
		{
			name:          "rejects invalid score",
			args:          []string{"zset_key", "not-a-float", "zset_member"},
			expectedError: "-ERR value is not a valid float\r\n",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			key, score, member, errorResponse := parseZaddCommandArguments(&RedisCommand{
				Type: CmdZADD,
				Args: testCase.args,
			})

			if errorResponse != testCase.expectedError {
				t.Errorf("parseZaddCommandArguments() error = %q, expected %q", errorResponse, testCase.expectedError)
			}

			if key != testCase.expectedKey {
				t.Errorf("parseZaddCommandArguments() key = %q, expected %q", key, testCase.expectedKey)
			}

			if score != testCase.expectedScore {
				t.Errorf("parseZaddCommandArguments() score = %v, expected %v", score, testCase.expectedScore)
			}

			if member != testCase.expectedMember {
				t.Errorf("parseZaddCommandArguments() member = %q, expected %q", member, testCase.expectedMember)
			}
		})
	}
}

func TestHandleZaddCodecraftersStageExample(t *testing.T) {
	resetZaddTestState(t)

	result := HandleZadd(&RedisCommand{
		Type: CmdZADD,
		Args: []string{"zset_key", "10.0", "zset_member"},
	})

	if result != ":1\r\n" {
		t.Errorf("HandleZadd() = %q, expected %q", result, ":1\r\n")
	}
}

func TestHandleZaddCreatesSortedSetWithMemberAndScore(t *testing.T) {
	resetZaddTestState(t)

	result := HandleZadd(&RedisCommand{
		Type: CmdZADD,
		Args: []string{"zset_key", "10.0", "zset_member"},
	})
	if result != ":1\r\n" {
		t.Fatalf("HandleZadd() = %q, expected %q", result, ":1\r\n")
	}

	sortedSet := GetInstance().GetSortedSet("zset_key")
	if sortedSet == nil {
		t.Fatal("expected sorted set to be created at key zset_key")
	}

	if sortedSet.MemberCount() != 1 {
		t.Fatalf("sorted set member count = %d, expected 1", sortedSet.MemberCount())
	}

	score, exists := sortedSet.GetMemberScore("zset_member")
	if !exists {
		t.Fatal("expected zset_member to exist in sorted set")
	}

	if score != 10.0 {
		t.Errorf("zset_member score = %v, expected 10.0", score)
	}
}

func TestHandleZaddArgumentParsing(t *testing.T) {
	resetZaddTestState(t)

	result := HandleZadd(&RedisCommand{
		Type: CmdZADD,
		Args: []string{"zset_key", "10.0"},
	})

	if result != "-ERR wrong number of arguments for 'zadd' command\r\n" {
		t.Errorf("HandleZadd() = %q, expected %q", result, "-ERR wrong number of arguments for 'zadd' command\r\n")
	}
}
