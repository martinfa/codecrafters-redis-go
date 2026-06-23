package main

import "math/rand/v2"

const (
	skipListMaxLevel    = 32
	skipListProbability = 0.25
)

type SkipListNode struct {
	member  string
	score   float64
	forward []*SkipListNode
}

type SkipList struct {
	head   *SkipListNode
	level  int
	length int
}

func NewSkipList() *SkipList {
	return &SkipList{
		head: &SkipListNode{
			forward: make([]*SkipListNode, skipListMaxLevel),
		},
		level: 1,
	}
}

func (skipList *SkipList) randomLevel() int {
	level := 1
	for level < skipListMaxLevel && rand.Float64() < skipListProbability {
		level++
	}
	return level
}

func (skipList *SkipList) Insert(score float64, member string) {
	predecessors := make([]*SkipListNode, skipListMaxLevel)
	current := skipList.head

	for levelIndex := skipList.level - 1; levelIndex >= 0; levelIndex-- {
		for current.forward[levelIndex] != nil &&
			compareSkipListMembers(current.forward[levelIndex].score, current.forward[levelIndex].member, score, member) < 0 {
			current = current.forward[levelIndex]
		}
		predecessors[levelIndex] = current
	}

	newLevel := skipList.randomLevel()
	node := &SkipListNode{
		member:  member,
		score:   score,
		forward: make([]*SkipListNode, newLevel),
	}

	if newLevel > skipList.level {
		for levelIndex := skipList.level; levelIndex < newLevel; levelIndex++ {
			predecessors[levelIndex] = skipList.head
		}
		skipList.level = newLevel
	}

	for levelIndex := 0; levelIndex < newLevel; levelIndex++ {
		node.forward[levelIndex] = predecessors[levelIndex].forward[levelIndex]
		predecessors[levelIndex].forward[levelIndex] = node
	}

	skipList.length++
}

func (skipList *SkipList) GetRank(member string) (rank int, found bool) {
	rank = 0
	current := skipList.head.forward[0]

	for current != nil {
		if current.member == member {
			return rank, true
		}
		rank++
		current = current.forward[0]
	}

	return 0, false
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
