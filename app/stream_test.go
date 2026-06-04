package main

import "testing"

func TestEntryIDGreaterThan(t *testing.T) {
	tests := []struct {
		name          string
		leftEntryID   string
		rightEntryID  string
		expectedGreater bool
	}{
		{name: "greater milliseconds", leftEntryID: "2-0", rightEntryID: "1-9", expectedGreater: true},
		{name: "equal milliseconds greater sequence", leftEntryID: "1-3", rightEntryID: "1-2", expectedGreater: true},
		{name: "equal id", leftEntryID: "1-2", rightEntryID: "1-2", expectedGreater: false},
		{name: "lower milliseconds despite higher sequence", leftEntryID: "0-3", rightEntryID: "1-2", expectedGreater: false},
		{name: "numeric not lexicographic compare", leftEntryID: "10-0", rightEntryID: "9-0", expectedGreater: true},
		{name: "first valid id after zero", leftEntryID: "0-1", rightEntryID: "0-0", expectedGreater: true},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := entryIDGreaterThan(testCase.leftEntryID, testCase.rightEntryID)
			if result != testCase.expectedGreater {
				t.Errorf("entryIDGreaterThan(%q, %q) = %v, expected %v",
					testCase.leftEntryID, testCase.rightEntryID, result, testCase.expectedGreater)
			}
		})
	}
}

func TestLastStreamEntryID(t *testing.T) {
	if lastStreamEntryID(nil) != zeroEntryID {
		t.Errorf("expected %q for nil stream, got %q", zeroEntryID, lastStreamEntryID(nil))
	}

	stream := &Stream{Entries: []StreamEntry{{ID: "1-2", Fields: map[string]string{}}}}
	if lastStreamEntryID(stream) != "1-2" {
		t.Errorf("expected last entry id 1-2, got %q", lastStreamEntryID(stream))
	}
}
