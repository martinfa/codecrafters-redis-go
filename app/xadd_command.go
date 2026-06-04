package main

import "fmt"

func HandleXadd(command *RedisCommand) string {
	if len(command.Args) < 4 || len(command.Args)%2 != 0 {
		return "-ERR wrong number of arguments for 'xadd' command\r\n"
	}

	streamKey := command.Args[0]
	entryID := command.Args[1]

	if entryID == zeroEntryID {
		return errXaddIDMustBeGreaterThanZeroZero
	}

	cache := GetInstance()
	stream := cache.GetStream(streamKey)
	if !entryIDGreaterThan(entryID, lastStreamEntryID(stream)) {
		return errXaddIDEqualOrSmallerThanTop
	}

	fields := make(map[string]string)
	for index := 2; index < len(command.Args); index += 2 {
		fields[command.Args[index]] = command.Args[index+1]
	}

	cache.AddStreamEntry(streamKey, entryID, fields)

	return fmt.Sprintf("$%d\r\n%s\r\n", len(entryID), entryID)
}
