package clock

import (
	"time"
)

// Clock is a time abstraction which can be used to if you need to mock time in tests.
type Clock interface {
	Now() time.Time
	NewTicker(d time.Duration) (<-chan time.Time, func())
	NewTimer(d time.Duration) (<-chan time.Time, func() bool)
	Since(t time.Time) time.Duration
}
