package main

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	errXaddIDMustBeGreaterThanZeroZero     = "-ERR The ID specified in XADD must be greater than 0-0\r\n"
	errXaddIDEqualOrSmallerThanTop         = "-ERR The ID specified in XADD is equal or smaller than the target stream top item\r\n"
	zeroEntryID                            = "0-0"
)

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

func parseEntryID(entryID string) (milliseconds int64, sequence int64, err error) {
	parts := strings.Split(entryID, "-")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid entry id %q", entryID)
	}

	milliseconds, err = strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, 0, err
	}

	sequence, err = strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, 0, err
	}

	return milliseconds, sequence, nil
}

func entryIDGreaterThan(leftEntryID string, rightEntryID string) bool {
	leftMilliseconds, leftSequence, err := parseEntryID(leftEntryID)
	if err != nil {
		return false
	}

	rightMilliseconds, rightSequence, err := parseEntryID(rightEntryID)
	if err != nil {
		return false
	}

	if leftMilliseconds > rightMilliseconds {
		return true
	}
	if leftMilliseconds < rightMilliseconds {
		return false
	}

	return leftSequence > rightSequence
}

func lastStreamEntryID(stream *Stream) string {
	if stream == nil || len(stream.Entries) == 0 {
		return zeroEntryID
	}

	return stream.Entries[len(stream.Entries)-1].ID
}
