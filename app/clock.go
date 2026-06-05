package main

import "time"

var CurrentTimeMilliseconds = func() int64 {
	return time.Now().UnixMilli()
}

func withFixedCurrentTimeMilliseconds(fixedMilliseconds int64, action func()) {
	originalCurrentTimeMilliseconds := CurrentTimeMilliseconds
	CurrentTimeMilliseconds = func() int64 {
		return fixedMilliseconds
	}
	defer func() {
		CurrentTimeMilliseconds = originalCurrentTimeMilliseconds
	}()

	action()
}
