package clock

import "time"

// Simple abstraction to allow for time based assertions in tests
type Clock interface {
	GetTime() int64
}

func New() Clock {
	return &clock{}
}

type clock struct{}

func (c clock) GetTime() int64 {
	return time.Now().UnixMilli()
}

type MockClock struct {
	time int64
}

func (c *MockClock) GetTime() int64 {
	return c.time
}

func (c *MockClock) SetTime(time int64) {
	c.time = time
}
