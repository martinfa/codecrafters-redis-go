package main

import "math/rand/v2"

const (
	skipListMaxLevel    = 32
	skipListProbability = 0.25
	groundLevel         = 0 // level 0 = the full sorted chain of all members
)

type SkipListNode struct {
	member  string
	score   float64
	forward []*SkipListNode
	span    []int // spanToNextAtLevel[i] = ground-chain members skipped by forward[i]
}

type SkipList struct {
	head   *SkipListNode
	level  int
	length int

	// testLevelOverride forces randomLevel() to return this value (tests only).
	testLevelOverride int
}

func NewSkipList() *SkipList {
	return &SkipList{
		head: &SkipListNode{
			forward: make([]*SkipListNode, skipListMaxLevel),
			span:    make([]int, skipListMaxLevel),
		},
		level: 1,
	}
}

func (node *SkipListNode) nextAtLevel(level int) *SkipListNode {
	return node.forward[level]
}

func (node *SkipListNode) setNextAtLevel(level int, next *SkipListNode) {
	node.forward[level] = next
}

func (node *SkipListNode) spanToNextAtLevel(level int) int {
	return node.span[level]
}

func (node *SkipListNode) setSpanToNextAtLevel(level int, span int) {
	node.span[level] = span
}

func (node *SkipListNode) nextOnGroundChain() *SkipListNode {
	return node.forward[groundLevel]
}

func (skipList *SkipList) randomLevel() int {
	if skipList.testLevelOverride > 0 {
		return skipList.testLevelOverride
	}
	level := 1
	for level < skipListMaxLevel && rand.Float64() < skipListProbability {
		level++
	}
	return level
}

func (skipList *SkipList) Insert(score float64, member string) {
	nodeBeforeInsertAtLevel := make([]*SkipListNode, skipListMaxLevel)
	sortedPositionAtLevel := make([]int, skipListMaxLevel)
	current := skipList.head
	topLevel := skipList.level - 1

	for searchLevel := topLevel; searchLevel >= groundLevel; searchLevel-- {
		if searchLevel == topLevel {
			sortedPositionAtLevel[searchLevel] = 0
		} else {
			sortedPositionAtLevel[searchLevel] = sortedPositionAtLevel[searchLevel+1]
		}

		for {
			nextNode := current.nextAtLevel(searchLevel)
			if nextNode == nil {
				break
			}
			nextSortsBeforeInsert := compareSkipListMembers(nextNode.score, nextNode.member, score, member) < 0
			if !nextSortsBeforeInsert {
				break
			}

			groundChainMembersSkipped := current.spanToNextAtLevel(searchLevel)
			sortedPositionAtLevel[searchLevel] += groundChainMembersSkipped
			current = nextNode
		}

		nodeBeforeInsertAtLevel[searchLevel] = current
	}

	newNodeLevel := skipList.randomLevel()
	newNode := &SkipListNode{
		member:  member,
		score:   score,
		forward: make([]*SkipListNode, newNodeLevel),
		span:    make([]int, newNodeLevel),
	}

	if newNodeLevel > skipList.level {
		for searchLevel := skipList.level; searchLevel < newNodeLevel; searchLevel++ {
			nodeBeforeInsertAtLevel[searchLevel] = skipList.head
			sortedPositionAtLevel[searchLevel] = 0
			skipList.head.setSpanToNextAtLevel(searchLevel, skipList.length)
		}
		skipList.level = newNodeLevel
	}

	insertSortedPosition := sortedPositionAtLevel[groundLevel]

	for spliceLevel := groundLevel; spliceLevel < newNodeLevel; spliceLevel++ {
		nodeBeforeInsert := nodeBeforeInsertAtLevel[spliceLevel]
		oldNext := nodeBeforeInsert.nextAtLevel(spliceLevel)

		newNode.setNextAtLevel(spliceLevel, oldNext)
		nodeBeforeInsert.setNextAtLevel(spliceLevel, newNode)

		sortedPositionAtBookmark := sortedPositionAtLevel[spliceLevel]
		distanceFromBookmarkToInsert := insertSortedPosition - sortedPositionAtBookmark
		oldSpanFromBookmarkToOldNext := nodeBeforeInsert.spanToNextAtLevel(spliceLevel)

		spanFromBookmarkToNewNode := distanceFromBookmarkToInsert + 1
		spanFromNewNodeToOldNext := oldSpanFromBookmarkToOldNext - distanceFromBookmarkToInsert

		nodeBeforeInsert.setSpanToNextAtLevel(spliceLevel, spanFromBookmarkToNewNode)
		newNode.setSpanToNextAtLevel(spliceLevel, spanFromNewNodeToOldNext)
	}

	for untouchedLevel := newNodeLevel; untouchedLevel < skipList.level; untouchedLevel++ {
		nodeBeforeInsert := nodeBeforeInsertAtLevel[untouchedLevel]
		spanBeforeInsert := nodeBeforeInsert.spanToNextAtLevel(untouchedLevel)
		nodeBeforeInsert.setSpanToNextAtLevel(untouchedLevel, spanBeforeInsert+1)
	}

	skipList.length++
}

func (skipList *SkipList) GetRank(score float64, member string) (memberRank int, found bool) {
	memberRank = 0
	current := skipList.head
	topLevel := skipList.level - 1

	for searchLevel := topLevel; searchLevel >= groundLevel; searchLevel-- {
		for {
			nextNode := current.nextAtLevel(searchLevel)
			if nextNode == nil {
				break
			}
			nextSortsBeforeTarget := compareSkipListMembers(nextNode.score, nextNode.member, score, member) < 0
			if !nextSortsBeforeTarget {
				break
			}

			groundChainMembersSkipped := current.spanToNextAtLevel(searchLevel)
			memberRank += groundChainMembersSkipped
			current = nextNode
		}
	}

	targetNode := current.nextOnGroundChain()
	targetFound := targetNode != nil && targetNode.member == member && targetNode.score == score
	if targetFound {
		return memberRank, true
	}

	return 0, false
}

func (skipList *SkipList) getNodeByRank(memberRank int) *SkipListNode {
	targetSortedPosition := memberRank + 1 // span math uses 1-based distance from head
	traversedGroundChainMembers := 0
	current := skipList.head
	topLevel := skipList.level - 1

	for searchLevel := topLevel; searchLevel >= groundLevel; searchLevel-- {
		for {
			nextNode := current.nextAtLevel(searchLevel)
			if nextNode == nil {
				break
			}

			spanToNext := current.spanToNextAtLevel(searchLevel)
			wouldOvershoot := traversedGroundChainMembers+spanToNext > targetSortedPosition
			if wouldOvershoot {
				break
			}

			traversedGroundChainMembers += spanToNext
			current = nextNode
		}

		if traversedGroundChainMembers == targetSortedPosition {
			return current
		}
	}

	return nil
}

func (skipList *SkipList) RangeByRank(startIndex int, stopIndex int) []string {
	if startIndex < 0 || startIndex >= skipList.length || startIndex > stopIndex {
		return []string{}
	}
	if stopIndex >= skipList.length {
		stopIndex = skipList.length - 1
	}

	members := make([]string, 0, stopIndex-startIndex+1)
	current := skipList.getNodeByRank(startIndex)

	for memberRank := startIndex; memberRank <= stopIndex && current != nil; memberRank++ {
		members = append(members, current.member)
		current = current.nextOnGroundChain()
	}

	return members
}

func (skipList *SkipList) Length() int {
	return skipList.length
}

func compareSkipListMembers(leftScore float64, leftMember string, rightScore float64, rightMember string) int {
	if leftScore < rightScore {
		return -1
	}
	if leftScore > rightScore {
		return 1
	}
	if leftMember < rightMember {
		return -1
	}
	if leftMember > rightMember {
		return 1
	}
	return 0
}
