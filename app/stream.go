package main

type StreamEntry struct {
	ID     string
	Fields map[string]string
}

type Stream struct {
	Entries []StreamEntry
}

func (c *Cache) AddStreamEntry(streamKey string, entryID string, fields map[string]string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	var stream *Stream
	item, exists := c.cache[streamKey]
	if exists {
		stream, _ = item.Value.(*Stream)
	}

	if stream == nil {
		stream = &Stream{Entries: []StreamEntry{}}
	}

	stream.Entries = append(stream.Entries, StreamEntry{
		ID:     entryID,
		Fields: fields,
	})

	c.cache[streamKey] = CacheItem{
		Value:      stream,
		Expiration: 0,
	}
}

func (c *Cache) GetStream(streamKey string) *Stream {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	item, exists := c.cache[streamKey]
	if !exists {
		return nil
	}

	stream, ok := item.Value.(*Stream)
	if !ok {
		return nil
	}

	return stream
}
