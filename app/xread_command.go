package main

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const blockingXreadTimeoutResponse = "*-1\r\n"

func parseXreadStreamsArguments(args []string) (streamKeys []string, startIDs []string, errorResponse string) {
	if len(args) < 3 {
		return nil, nil, "-ERR wrong number of arguments for 'xread' command\r\n"
	}

	if !strings.EqualFold(args[0], "STREAMS") {
		return nil, nil, "-ERR wrong number of arguments for 'xread' command\r\n"
	}

	streamKeysAndIds := args[1:]
	if len(streamKeysAndIds)%2 != 0 {
		return nil, nil, "-ERR wrong number of arguments for 'xread' command\r\n"
	}

	streamCount := len(streamKeysAndIds) / 2
	streamKeys = make([]string, streamCount)
	startIDs = make([]string, streamCount)
	for index := 0; index < streamCount; index++ {
		streamKeys[index] = streamKeysAndIds[index]
		startIDs[index] = streamKeysAndIds[streamCount+index]
	}

	return streamKeys, startIDs, ""
}

func parseXreadCommand(command *RedisCommand) (streamKeys []string, startIDs []string, blockMilliseconds int, errorResponse string) {
	blockMilliseconds = -1
	args := command.Args

	if len(args) >= 2 && strings.EqualFold(args[0], "BLOCK") {
		parsedBlockMilliseconds, parseError := strconv.Atoi(args[1])
		if parseError != nil {
			return nil, nil, -1, "-ERR wrong number of arguments for 'xread' command\r\n"
		}

		blockMilliseconds = parsedBlockMilliseconds
		args = args[2:]
	}

	streamKeys, startIDs, errorResponse = parseXreadStreamsArguments(args)
	return streamKeys, startIDs, blockMilliseconds, errorResponse
}

func parseXreadCommandArguments(command *RedisCommand) (streamKey string, startID string, errorResponse string) {
	streamKeys, startIDs, blockMilliseconds, errorResponse := parseXreadCommand(command)
	if errorResponse != "" {
		return "", "", errorResponse
	}

	if blockMilliseconds >= 0 {
		return "", "", "-ERR wrong number of arguments for 'xread' command\r\n"
	}

	if len(streamKeys) != 1 {
		return "", "", "-ERR wrong number of arguments for 'xread' command\r\n"
	}

	return streamKeys[0], startIDs[0], ""
}

func parseManyXReadCommandArguments(command *RedisCommand) (streamKeys []string, startIds []string, errorResponse string) {
	streamKeys, startIds, blockMilliseconds, errorResponse := parseXreadCommand(command)
	if errorResponse != "" {
		return nil, nil, errorResponse
	}

	if blockMilliseconds >= 0 {
		return nil, nil, "-ERR wrong number of arguments for 'xread' command\r\n"
	}

	if len(streamKeys) <= 1 {
		return nil, nil, "-ERR wrong number of arguments for 'xread' command\r\n"
	}

	return streamKeys, startIds, ""
}

func resolveXreadStartIDs(streamKeys []string, startIDs []string) []string {
	resolvedStartIDs := make([]string, len(startIDs))
	for index, startID := range startIDs {
		if startID == xreadNewEntriesSentinel {
			stream := GetInstance().GetStream(streamKeys[index])
			resolvedStartIDs[index] = lastStreamEntryID(stream)
			continue
		}

		resolvedStartIDs[index] = startID
	}

	return resolvedStartIDs
}

func buildXreadResponse(streamKey string, startID string) XreadResponse {
	startBoundID, err := normalizeRangeBoundID(startID, startSequenceNumberDefault)
	if err != nil {
		return XreadResponse{Streams: []XreadStreamResponse{}}
	}

	stream := GetInstance().GetStream(streamKey)
	entries := []XreadEntryResponse{}
	if stream != nil {
		for _, entry := range stream.Entries {
			if !entryIDAfterExclusiveStart(entry.ID, startBoundID) {
				continue
			}

			entries = append(entries, XreadEntryResponse{
				EntryID:     entry.ID,
				FieldValues: append([]string(nil), entry.FieldValues...),
			})
		}
	}

	if len(entries) == 0 {
		return XreadResponse{Streams: []XreadStreamResponse{}}
	}

	return XreadResponse{
		Streams: []XreadStreamResponse{{
			StreamKey: streamKey,
			Entries:   entries,
		}},
	}
}

func buildMultiXreadResponse(streamKeys []string, startIDs []string) XreadResponse {
	xreadResponse := XreadResponse{Streams: []XreadStreamResponse{}}
	for index := range streamKeys {
		streamResponse := buildXreadResponse(streamKeys[index], startIDs[index])
		if len(streamResponse.Streams) > 0 {
			xreadResponse.Streams = append(xreadResponse.Streams, streamResponse.Streams[0])
		}
	}

	return xreadResponse
}

func encodeXreadResponse(xreadResponse XreadResponse) string {
	if len(xreadResponse.Streams) == 0 {
		return "*0\r\n"
	}

	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("*%d\r\n", len(xreadResponse.Streams)))
	for _, stream := range xreadResponse.Streams {
		builder.WriteString("*2\r\n")
		builder.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(stream.StreamKey), stream.StreamKey))
		builder.WriteString(fmt.Sprintf("*%d\r\n", len(stream.Entries)))
		for _, entry := range stream.Entries {
			builder.WriteString("*2\r\n")
			builder.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(entry.EntryID), entry.EntryID))
			builder.WriteString(fmt.Sprintf("*%d\r\n", len(entry.FieldValues)))
			for _, fieldValue := range entry.FieldValues {
				builder.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(fieldValue), fieldValue))
			}
		}
	}

	return builder.String()
}

func waitForBlockingXreadResponse(streamKeys []string, startIDs []string, blockMilliseconds int) string {
	xreadResponse := buildMultiXreadResponse(streamKeys, startIDs)
	if len(xreadResponse.Streams) > 0 {
		return encodeXreadResponse(xreadResponse)
	}

	subscriptions := make([]Subscription, len(streamKeys))
	for index, streamKey := range streamKeys {
		subscriptions[index] = GetEventBus().Subscribe(EventStreamChanged, StreamKeyFilter(streamKey))
	}
	defer func() {
		for _, subscription := range subscriptions {
			subscription.Unregister()
		}
	}()

	var selectCases []reflect.SelectCase
	if blockMilliseconds > 0 {
		timer := time.NewTimer(time.Duration(blockMilliseconds) * time.Millisecond)
		defer timer.Stop()

		selectCases = make([]reflect.SelectCase, len(subscriptions)+1)
		selectCases[0] = reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(timer.C),
		}
		for index, subscription := range subscriptions {
			selectCases[index+1] = reflect.SelectCase{
				Dir:  reflect.SelectRecv,
				Chan: reflect.ValueOf(subscription.Notifications),
			}
		}
	} else {
		selectCases = make([]reflect.SelectCase, len(subscriptions))
		for index, subscription := range subscriptions {
			selectCases[index] = reflect.SelectCase{
				Dir:  reflect.SelectRecv,
				Chan: reflect.ValueOf(subscription.Notifications),
			}
		}
	}

	for {
		chosen, _, _ := reflect.Select(selectCases)
		if blockMilliseconds > 0 && chosen == 0 {
			return blockingXreadTimeoutResponse
		}

		xreadResponse = buildMultiXreadResponse(streamKeys, startIDs)
		if len(xreadResponse.Streams) > 0 {
			return encodeXreadResponse(xreadResponse)
		}
	}
}

func HandleXread(command *RedisCommand) string {
	streamKeys, startIDs, blockMilliseconds, errorResponse := parseXreadCommand(command)
	if errorResponse != "" {
		return errorResponse
	}

	startIDs = resolveXreadStartIDs(streamKeys, startIDs)

	if blockMilliseconds >= 0 {
		return waitForBlockingXreadResponse(streamKeys, startIDs, blockMilliseconds)
	}

	return encodeXreadResponse(buildMultiXreadResponse(streamKeys, startIDs))
}
