package main

import (
	"fmt"
	"testing"
)

func resetSortedSetTestState(t *testing.T) {
	t.Helper()
	GetInstance().cache = make(map[string]CacheItem)
}

func addSortedSetMembers(t *testing.T, key string, members ...SortedSetMember) {
	t.Helper()

	for _, member := range members {
		result := HandleZadd(&RedisCommand{
			Type: CmdZADD,
			Args: []string{key, formatFloatScore(member.Score), member.Member},
		})
		if result != ":1\r\n" {
			t.Fatalf("HandleZadd(%q, %v) = %q, expected %q", member.Member, member.Score, result, ":1\r\n")
		}
	}
}

func formatFloatScore(score float64) string {
	return fmt.Sprintf("%g", score)
}

func TestParseCommandType_ZRANKReturnsKnownCommand(t *testing.T) {
	commandType := ParseCommandType("ZRANK")
	if commandType != CmdZRANK {
		t.Fatalf("expected ZRANK to parse as CmdZRANK, got %v", commandType)
	}
}

func TestParseZrankCommandArguments(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedKey    string
		expectedMember string
		expectedError  string
	}{
		{
			name:           "parses key and member",
			args:           []string{"zset_key", "caz"},
			expectedKey:    "zset_key",
			expectedMember: "caz",
		},
		{
			name:          "rejects missing arguments",
			args:          []string{},
			expectedError: "-ERR wrong number of arguments for 'zrank' command\r\n",
		},
		{
			name:          "rejects key only",
			args:          []string{"zset_key"},
			expectedError: "-ERR wrong number of arguments for 'zrank' command\r\n",
		},
		{
			name:          "rejects too many arguments",
			args:          []string{"zset_key", "caz", "extra"},
			expectedError: "-ERR wrong number of arguments for 'zrank' command\r\n",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			key, member, errorResponse := parseZrankCommandArguments(&RedisCommand{
				Type: CmdZRANK,
				Args: testCase.args,
			})

			if errorResponse != testCase.expectedError {
				t.Errorf("parseZrankCommandArguments() error = %q, expected %q", errorResponse, testCase.expectedError)
			}

			if key != testCase.expectedKey {
				t.Errorf("parseZrankCommandArguments() key = %q, expected %q", key, testCase.expectedKey)
			}

			if member != testCase.expectedMember {
				t.Errorf("parseZrankCommandArguments() member = %q, expected %q", member, testCase.expectedMember)
			}
		})
	}
}

func TestHandleZrankStageDocumentationExample(t *testing.T) {
	resetSortedSetTestState(t)

	addSortedSetMembers(t, "zset_key",
		SortedSetMember{Member: "member_with_score_1", Score: 1.0},
		SortedSetMember{Member: "member_with_score_2", Score: 2.0},
		SortedSetMember{Member: "another_member_with_score_2", Score: 2.0},
	)

	tests := []struct {
		member       string
		expectedResp string
	}{
		{member: "member_with_score_1", expectedResp: ":0\r\n"},
		{member: "another_member_with_score_2", expectedResp: ":1\r\n"},
		{member: "member_with_score_2", expectedResp: ":2\r\n"},
	}

	for _, testCase := range tests {
		result := HandleZrank(&RedisCommand{
			Type: CmdZRANK,
			Args: []string{"zset_key", testCase.member},
		})
		if result != testCase.expectedResp {
			t.Errorf("HandleZrank(%q) = %q, expected %q", testCase.member, result, testCase.expectedResp)
		}
	}
}

func TestHandleZrankCodecraftersStageExample(t *testing.T) {
	resetSortedSetTestState(t)

	addSortedSetMembers(t, "zset_key",
		SortedSetMember{Member: "foo", Score: 100.0},
		SortedSetMember{Member: "bar", Score: 100.0},
		SortedSetMember{Member: "baz", Score: 20.0},
		SortedSetMember{Member: "caz", Score: 30.1},
		SortedSetMember{Member: "paz", Score: 40.2},
	)

	tests := []struct {
		member       string
		expectedResp string
	}{
		{member: "caz", expectedResp: ":1\r\n"},
		{member: "baz", expectedResp: ":0\r\n"},
		{member: "foo", expectedResp: ":4\r\n"},
		{member: "bar", expectedResp: ":3\r\n"},
	}

	for _, testCase := range tests {
		result := HandleZrank(&RedisCommand{
			Type: CmdZRANK,
			Args: []string{"zset_key", testCase.member},
		})
		if result != testCase.expectedResp {
			t.Errorf("HandleZrank(%q) = %q, expected %q", testCase.member, result, testCase.expectedResp)
		}
	}
}

func TestHandleZrankMissingMemberReturnsNullBulkString(t *testing.T) {
	resetSortedSetTestState(t)

	addSortedSetMembers(t, "zset_key", SortedSetMember{Member: "foo", Score: 1.0})

	result := HandleZrank(&RedisCommand{
		Type: CmdZRANK,
		Args: []string{"zset_key", "missing_member"},
	})
	if result != "$-1\r\n" {
		t.Errorf("HandleZrank() = %q, expected %q", result, "$-1\r\n")
	}
}

func TestHandleZrankMissingKeyReturnsNullBulkString(t *testing.T) {
	resetSortedSetTestState(t)

	result := HandleZrank(&RedisCommand{
		Type: CmdZRANK,
		Args: []string{"missing_key", "member"},
	})
	if result != "$-1\r\n" {
		t.Errorf("HandleZrank() = %q, expected %q", result, "$-1\r\n")
	}
}

func TestHandleZrankArgumentParsing(t *testing.T) {
	resetSortedSetTestState(t)

	result := HandleZrank(&RedisCommand{
		Type: CmdZRANK,
		Args: []string{"zset_key"},
	})
	if result != "-ERR wrong number of arguments for 'zrank' command\r\n" {
		t.Errorf("HandleZrank() = %q, expected %q", result, "-ERR wrong number of arguments for 'zrank' command\r\n")
	}
}
