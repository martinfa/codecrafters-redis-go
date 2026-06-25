package main

import "testing"

// TestSkipListSpanDebugWalkthrough is the best test to step-debug.
//
// Set breakpoints on the lines marked BREAKPOINT below, then run:
//
//	go test ./app/... -run TestSkipListSpanDebugWalkthrough -count=1
//
// Phase 1 — Insert baz: first node, trivial spans
// Phase 2 — Insert caz: middle insert, span split (the key moment)
// Phase 3 — Insert paz: append at end
// Phase 4 — GetRank(caz): read spans back
// Phase 5 — RangeByRank(1,2): jump to rank via spans
func TestSkipListSpanDebugWalkthrough(t *testing.T) {
	skipList := NewSkipList()
	skipList.testLevelOverride = 1 // every node is level-0 only → spans are always 1

	// BREAKPOINT: step into Insert
	skipList.Insert(20.0, "baz")
	// After: HEAD.forward[0]=baz, HEAD.span[0]=1, baz.span[0]=0

	// BREAKPOINT: step into Insert — watch rank[0] become 1, span split
	skipList.Insert(30.1, "caz")
	// After: baz→caz→paz chain, each span[0]=1

	// BREAKPOINT: step into Insert — insert at tail
	skipList.Insert(40.2, "paz")
	// After: baz(rank0) → caz(rank1) → paz(rank2)

	// BREAKPOINT: step into GetRank — rank += span as you hop
	rank, found := skipList.GetRank(30.1, "caz")
	if !found || rank != 1 {
		t.Fatalf("GetRank(caz) = (%d, %v), want (1, true)", rank, found)
	}

	// BREAKPOINT: step into RangeByRank → getNodeByRank
	members := skipList.RangeByRank(1, 2)
	if len(members) != 2 || members[0] != "caz" || members[1] != "paz" {
		t.Fatalf("RangeByRank(1,2) = %v, want [caz paz]", members)
	}
}
