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
