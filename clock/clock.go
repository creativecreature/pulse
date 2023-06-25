// Package clock is a simple abstraction to allow for time based assertions in tests
package clock

import "time"

type Clock struct{}

func New() Clock {
	return Clock{}
}

// GetTime returns the current UTC time in milliseconds.
func (c Clock) GetTime() int64 {
	return time.Now().UTC().UnixMilli()
}
