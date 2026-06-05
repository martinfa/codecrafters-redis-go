package main

import (
	"fmt"
	"strings"
)

func parseXreadCommandArguments(command *RedisCommand) (streamKey string, startID string, errorResponse string) {
	if len(command.Args) != 3 {
		return "", "", "-ERR wrong number of arguments for 'xread' command\r\n"
	}

	if !strings.EqualFold(command.Args[0], "STREAMS") {
		return "", "", "-ERR wrong number of arguments for 'xread' command\r\n"
	}

	return command.Args[1], command.Args[2], ""
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

func HandleXread(command *RedisCommand) string {
	streamKey, startID, errorResponse := parseXreadCommandArguments(command)
	if errorResponse != "" {
		return errorResponse
	}

	return encodeXreadResponse(buildXreadResponse(streamKey, startID))
}
