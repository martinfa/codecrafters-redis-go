package main

import (
	"os"
	"strings"
	"testing"
)

func TestSkipListWalkthroughCoversCorePaths(t *testing.T) {
	content, err := os.ReadFile("../skip_list_walkthrough.html")
	if err != nil {
		t.Fatalf("read skip_list_walkthrough.html: %v", err)
	}

	requiredMarkers := []string{
		"data-case=\"insert-front\"",
		"data-case=\"insert-middle\"",
		"data-case=\"insert-tail\"",
		"data-case=\"insert-tall-node\"",
		"data-case=\"insert-short-node\"",
		"data-case=\"get-rank\"",
		"data-case=\"range-by-rank\"",
		"spanFromBookmarkToNewNode",
		"spanFromNewNodeToOldNext",
	}

	for _, requiredMarker := range requiredMarkers {
		if !strings.Contains(string(content), requiredMarker) {
			t.Fatalf("walkthrough missing marker %q", requiredMarker)
		}
	}
}
