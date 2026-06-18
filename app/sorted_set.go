package main

type SortedSetMember struct {
	Member string
	Score  float64
}

type SortedSet struct {
	memberScores   map[string]float64
	orderedMembers []SortedSetMember
}

func newSortedSet() *SortedSet {
	return &SortedSet{
		memberScores:   make(map[string]float64),
		orderedMembers: []SortedSetMember{},
	}
}

func (sortedSet *SortedSet) GetMemberScore(member string) (float64, bool) {
	score, exists := sortedSet.memberScores[member]
	return score, exists
}

func (sortedSet *SortedSet) MemberCount() int {
	return len(sortedSet.memberScores)
}

func (cache *Cache) GetSortedSet(key string) *SortedSet {
	value := cache.Get(key)
	if value == nil {
		return nil
	}

	sortedSet, _ := value.(*SortedSet)
	return sortedSet
}

func (cache *Cache) Zadd(key string, score float64, member string) int {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	var sortedSet *SortedSet
	item, exists := cache.cache[key]
	if exists {
		sortedSet, _ = item.Value.(*SortedSet)
	}

	if sortedSet == nil {
		sortedSet = newSortedSet()
	}

	if _, memberExists := sortedSet.memberScores[member]; memberExists {
		return 0
	}

	sortedSet.memberScores[member] = score
	sortedSet.orderedMembers = insertSortedSetMember(sortedSet.orderedMembers, SortedSetMember{
		Member: member,
		Score:  score,
	})

	cache.cache[key] = CacheItem{
		Value:      sortedSet,
		Expiration: 0,
	}

	return 1
}

func insertSortedSetMember(orderedMembers []SortedSetMember, newMember SortedSetMember) []SortedSetMember {
	insertIndex := len(orderedMembers)
	for index, existingMember := range orderedMembers {
		if newMember.Score < existingMember.Score {
			insertIndex = index
			break
		}
	}

	orderedMembers = append(orderedMembers, SortedSetMember{})
	copy(orderedMembers[insertIndex+1:], orderedMembers[insertIndex:])
	orderedMembers[insertIndex] = newMember

	return orderedMembers
}
