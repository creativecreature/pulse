package clock

import "time"

// RealClock imlements the Clock interface using the standard libraries time package.
type RealClock struct{}

func New() *RealClock {
	return &RealClock{}
}

// Now is a wrapper around time.Now().
func (c *RealClock) Now() time.Time {
	return time.Now()
}

// NewTicker is a wrapper around time.NewTicker().
func (c *RealClock) NewTicker(d time.Duration) (<-chan time.Time, func()) {
	t := time.NewTicker(d)
	return t.C, t.Stop
}

// NewTimer is a wrapper around time.NewTimer().
func (c *RealClock) NewTimer(d time.Duration) (<-chan time.Time, func() bool) {
	t := time.NewTimer(d)
	return t.C, t.Stop
}

// Since is a wrapper around time.Since().
func (c *RealClock) Since(t time.Time) time.Duration {
	return time.Since(t)
}
