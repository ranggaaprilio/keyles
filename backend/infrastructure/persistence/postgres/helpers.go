package postgres

import (
	"time"
)

// Helper function to convert Unix milliseconds to time.Time
func timeFromUnixMilli(ms int64) time.Time {
	return time.Unix(0, ms*int64(time.Millisecond))
}

// Helper function to convert time.Time to Unix milliseconds
func timeToUnixMilli(t time.Time) int64 {
	return t.UnixMilli()
}
