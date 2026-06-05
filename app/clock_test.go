package main

import (
	"testing"
	"time"
)

func TestCurrentTimeMillisecondsReturnsIncreasingValues(t *testing.T) {
	first := CurrentTimeMilliseconds()
	time.Sleep(1 * time.Millisecond)
	second := CurrentTimeMilliseconds()

	if second < first {
		t.Fatalf("expected time to advance, got first=%d second=%d", first, second)
	}
}

func TestCurrentTimeMillisecondsCanBeOverridden(t *testing.T) {
	originalCurrentTimeMilliseconds := CurrentTimeMilliseconds
	t.Cleanup(func() {
		CurrentTimeMilliseconds = originalCurrentTimeMilliseconds
	})

	CurrentTimeMilliseconds = func() int64 {
		return 1526919030474
	}

	if got := CurrentTimeMilliseconds(); got != 1526919030474 {
		t.Fatalf("CurrentTimeMilliseconds() = %d, expected 1526919030474", got)
	}
}

func TestWithFixedCurrentTimeMillisecondsRestoresOriginal(t *testing.T) {
	withFixedCurrentTimeMilliseconds(99, func() {
		if got := CurrentTimeMilliseconds(); got != 99 {
			t.Fatalf("inside override CurrentTimeMilliseconds() = %d, expected 99", got)
		}
	})

	if got := CurrentTimeMilliseconds(); got == 99 {
		t.Fatal("expected original clock to be restored after withFixedCurrentTimeMilliseconds")
	}
}
