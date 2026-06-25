package main

type SortedSetMember struct {
	Member string
	Score  float64
}

type SortedSet struct {
	memberScores map[string]float64
	orderedIndex *SkipList
}

func newSortedSet() *SortedSet {
	return &SortedSet{
		memberScores: make(map[string]float64),
		orderedIndex: NewSkipList(),
	}
}

func (sortedSet *SortedSet) GetMemberScore(member string) (float64, bool) {
	score, exists := sortedSet.memberScores[member]
	return score, exists
}

func (sortedSet *SortedSet) GetMemberRank(member string) (int, bool) {
	score, exists := sortedSet.memberScores[member]
	if !exists {
		return 0, false
	}
	return sortedSet.orderedIndex.GetRank(score, member)
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
	sortedSet.orderedIndex.Insert(score, member)

	cache.cache[key] = CacheItem{
		Value:      sortedSet,
		Expiration: 0,
	}

	return 1
}

func (sortedSet *SortedSet) GetMembersInRankRange(startIndex int, stopIndex int) []string {
	return sortedSet.orderedIndex.RangeByRank(startIndex, stopIndex)
}

func (cache *Cache) Zrange(key string, startIndex int, stopIndex int) []string {
	sortedSet := cache.GetSortedSet(key)
	if sortedSet == nil {
		return nil
	}

	return sortedSet.GetMembersInRankRange(startIndex, stopIndex)
}

func (cache *Cache) Zcard(key string) int {
	sortedSet := cache.GetSortedSet(key)
	if sortedSet == nil {
		return 0
	}

	return sortedSet.MemberCount()
}

func (cache *Cache) Zrank(key string, member string) (int, bool) {
	sortedSet := cache.GetSortedSet(key)
	if sortedSet == nil {
		return 0, false
	}

	return sortedSet.GetMemberRank(member)
}
