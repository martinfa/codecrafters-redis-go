package main

type List struct {
	Elements []string
}

func (cache *Cache) PushListRight(listKey string, elements ...string) int {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	var list *List
	item, exists := cache.cache[listKey]
	if exists {
		list, _ = item.Value.(*List)
	}

	if list == nil {
		list = &List{Elements: []string{}}
	}

	list.Elements = append(list.Elements, elements...)
	cache.cache[listKey] = CacheItem{
		Value:      list,
		Expiration: 0,
	}

	return len(list.Elements)
}

func (cache *Cache) PushListLeft(listKey string, elements ...string) int {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	var list *List
	item, exists := cache.cache[listKey]
	if exists {
		list, _ = item.Value.(*List)
	}

	if list == nil {
		list = &List{Elements: []string{}}
	}

	for _, element := range elements {
		list.Elements = append([]string{element}, list.Elements...)
	}

	cache.cache[listKey] = CacheItem{
		Value:      list,
		Expiration: 0,
	}

	return len(list.Elements)
}

func (cache *Cache) GetList(listKey string) *List {
	value := cache.Get(listKey)
	if value == nil {
		return nil
	}

	list, _ := value.(*List)
	return list
}

func (cache *Cache) GetListLength(listKey string) int {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	var list *List
	item, exists := cache.cache[listKey]
	if exists {
		list, _ = item.Value.(*List)
	}

	if list == nil {
		return 0
	}

	return len(list.Elements)
}

func (cache *Cache) PopListLeft(listKey string, count int) ([]string, bool) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	var list *List
	item, exists := cache.cache[listKey]
	if exists {
		list, _ = item.Value.(*List)
	}

	if list == nil || len(list.Elements) == 0 {
		return nil, false
	}

	if count > len(list.Elements) {
		count = len(list.Elements)
	}

	poppedElements := append([]string(nil), list.Elements[:count]...)
	list.Elements = list.Elements[count:]

	cache.cache[listKey] = CacheItem{
		Value:      list,
		Expiration: 0,
	}

	return poppedElements, true
}
