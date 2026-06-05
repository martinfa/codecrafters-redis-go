package main

import (
	"fmt"
	"strings"
)

func parseXrangeCommandArguments(command *RedisCommand) (streamKey string, startID string, endID string, errorResponse string) {
	if len(command.Args) != 3 {
		return "", "", "", "-ERR wrong number of arguments for 'xrange' command\r\n"
	}

	return command.Args[0], command.Args[1], command.Args[2], ""
}

func buildXrangeResponse(stream *Stream, startBoundID string, endBoundID string) XrangeResponse {
	xrangeResponse := XrangeResponse{Entries: []XrangeEntryResponse{}}
	if stream == nil {
		return xrangeResponse
	}

	for _, entry := range stream.Entries {
		if !entryIDInInclusiveRange(entry.ID, startBoundID, endBoundID) {
			continue
		}

		xrangeResponse.Entries = append(xrangeResponse.Entries, XrangeEntryResponse{
			EntryID:     entry.ID,
			FieldValues: append([]string(nil), entry.FieldValues...),
		})
	}

	return xrangeResponse
}

func encodeXrangeResponse(xrangeResponse XrangeResponse) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("*%d\r\n", len(xrangeResponse.Entries)))
	for _, entry := range xrangeResponse.Entries {
		builder.WriteString("*2\r\n")
		builder.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(entry.EntryID), entry.EntryID))
		builder.WriteString(fmt.Sprintf("*%d\r\n", len(entry.FieldValues)))
		for _, fieldValue := range entry.FieldValues {
			builder.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(fieldValue), fieldValue))
		}
	}

	return builder.String()
}

func HandleXrange(command *RedisCommand) string {
	streamKey, startID, endID, errorResponse := parseXrangeCommandArguments(command)
	if errorResponse != "" {
		return errorResponse
	}

	startBoundID, err := normalizeRangeBoundID(startID, startSequenceNumberDefault)
	if err != nil {
		return "*0\r\n"
	}

	stream := GetInstance().GetStream(streamKey)

	endBoundID, err := resolveXrangeEndBoundID(stream, endID)
	if err != nil {
		return "*0\r\n"
	}

	xrangeResponse := buildXrangeResponse(stream, startBoundID, endBoundID)

	return encodeXrangeResponse(xrangeResponse)
}
