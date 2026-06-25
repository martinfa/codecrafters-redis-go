package main

import "testing"

func newTestSkipList() *SkipList {
	return NewSkipList()
}

func insertSkipListMembers(skipList *SkipList, members ...SortedSetMember) {
	for _, member := range members {
		skipList.Insert(member.Score, member.Member)
	}
}

func TestCompareSkipListMembersOrdersByScoreThenMember(t *testing.T) {
	if compareSkipListMembers(1.0, "b", 2.0, "a") >= 0 {
		t.Fatal("expected lower score to sort first")
	}

	if compareSkipListMembers(2.0, "a", 2.0, "b") >= 0 {
		t.Fatal("expected lexicographically smaller member to sort first for tied scores")
	}
}

func TestNewSkipListHasZeroLength(t *testing.T) {
	skipList := newTestSkipList()

	if skipList.Length() != 0 {
		t.Fatalf("Length() = %d, expected 0", skipList.Length())
	}
}

func TestSkipListInsertIncreasesLength(t *testing.T) {
	skipList := newTestSkipList()

	skipList.Insert(10.0, "zset_member")

	if skipList.Length() != 1 {
		t.Fatalf("Length() = %d, expected 1", skipList.Length())
	}
}

func TestSkipListInsertOrdersByIncreasingScore(t *testing.T) {
	skipList := newTestSkipList()

	insertSkipListMembers(skipList,
		SortedSetMember{Member: "paz", Score: 40.2},
		SortedSetMember{Member: "baz", Score: 20.0},
		SortedSetMember{Member: "caz", Score: 30.1},
	)

	expectedRanks := map[string]struct {
		score float64
		rank  int
	}{
		"baz": {score: 20.0, rank: 0},
		"caz": {score: 30.1, rank: 1},
		"paz": {score: 40.2, rank: 2},
	}

	for member, expected := range expectedRanks {
		rank, found := skipList.GetRank(expected.score, member)
		if !found {
			t.Fatalf("GetRank(%q) member not found", member)
		}
		if rank != expected.rank {
			t.Errorf("GetRank(%q) = %d, expected %d", member, rank, expected.rank)
		}
	}
}

func TestSkipListInsertBreaksTiesLexicographically(t *testing.T) {
	skipList := newTestSkipList()

	insertSkipListMembers(skipList,
		SortedSetMember{Member: "member_with_score_1", Score: 1.0},
		SortedSetMember{Member: "member_with_score_2", Score: 2.0},
		SortedSetMember{Member: "another_member_with_score_2", Score: 2.0},
	)

	tests := []struct {
		member       string
		score        float64
		expectedRank int
	}{
		{member: "member_with_score_1", score: 1.0, expectedRank: 0},
		{member: "another_member_with_score_2", score: 2.0, expectedRank: 1},
		{member: "member_with_score_2", score: 2.0, expectedRank: 2},
	}

	for _, testCase := range tests {
		rank, found := skipList.GetRank(testCase.score, testCase.member)
		if !found {
			t.Fatalf("GetRank(%q) member not found", testCase.member)
		}
		if rank != testCase.expectedRank {
			t.Errorf("GetRank(%q) = %d, expected %d", testCase.member, rank, testCase.expectedRank)
		}
	}
}

func TestSkipListGetRankCodecraftersStageExample(t *testing.T) {
	skipList := newTestSkipList()

	insertSkipListMembers(skipList,
		SortedSetMember{Member: "foo", Score: 100.0},
		SortedSetMember{Member: "bar", Score: 100.0},
		SortedSetMember{Member: "baz", Score: 20.0},
		SortedSetMember{Member: "caz", Score: 30.1},
		SortedSetMember{Member: "paz", Score: 40.2},
	)

	tests := []struct {
		member       string
		score        float64
		expectedRank int
	}{
		{member: "baz", score: 20.0, expectedRank: 0},
		{member: "caz", score: 30.1, expectedRank: 1},
		{member: "paz", score: 40.2, expectedRank: 2},
		{member: "bar", score: 100.0, expectedRank: 3},
		{member: "foo", score: 100.0, expectedRank: 4},
	}

	for _, testCase := range tests {
		rank, found := skipList.GetRank(testCase.score, testCase.member)
		if !found {
			t.Fatalf("GetRank(%q) member not found", testCase.member)
		}
		if rank != testCase.expectedRank {
			t.Errorf("GetRank(%q) = %d, expected %d", testCase.member, rank, testCase.expectedRank)
		}
	}
}

func TestSkipListGetRankMissingMemberReturnsNotFound(t *testing.T) {
	skipList := newTestSkipList()
	insertSkipListMembers(skipList, SortedSetMember{Member: "foo", Score: 1.0})

	_, found := skipList.GetRank(0.0, "missing_member")
	if found {
		t.Fatal("GetRank(missing_member) found = true, expected false")
	}
}
