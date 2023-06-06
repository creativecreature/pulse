package clock

import "time"

func New() clock {
	return clock{}
}

type clock struct{}

func (c clock) GetTime() int64 {
	return time.Now().UTC().UnixMilli()
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
