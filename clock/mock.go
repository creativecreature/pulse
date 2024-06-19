package clock

import (
	"sync"
	"sync/atomic"
	"time"
)

type mockTimer struct {
	deadline time.Time
	ch       chan time.Time
	stopped  *atomic.Bool
}

type mockTicker struct {
	nextTick time.Time
	interval time.Duration
	ch       chan time.Time
	stopped  *atomic.Bool
}

// MockClock is a mock implementation of the Clock interface.
type MockClock struct {
	mu      sync.Mutex
	time    time.Time
	timers  []*mockTimer
	tickers []*mockTicker
}

func NewMock(time time.Time) *MockClock {
	var c MockClock
	c.time = time
	c.timers = make([]*mockTimer, 0)
	c.tickers = make([]*mockTicker, 0)
	return &c
}

// Set sets the time of the clock and triggers any timers or tickers that should fire.
func (c *MockClock) Set(t time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if t.Before(c.time) {
		panic("can't go back in time")
	}

	c.time = t
	for _, ticker := range c.tickers {
		if !ticker.stopped.Load() && !ticker.nextTick.Add(ticker.interval).After(c.time) {
			//nolint: durationcheck // This is a test clock, we don't care about overflows.
			nextTick := (c.time.Sub(ticker.nextTick) / ticker.interval) * ticker.interval
			ticker.nextTick = ticker.nextTick.Add(nextTick)
			select {
			case ticker.ch <- c.time:
			default:
			}
		}
	}

	unfiredTimers := make([]*mockTimer, 0)
	for i, timer := range c.timers {
		if timer.deadline.After(c.time) && !timer.stopped.Load() {
			unfiredTimers = append(unfiredTimers, c.timers[i])
			continue
		}
		timer.stopped.Store(true)
		timer.ch <- c.time
	}
	c.timers = unfiredTimers
}

// Add advances the clock by the duration and triggers any timers or tickers that should fire.
func (c *MockClock) Add(d time.Duration) {
	c.Set(c.time.Add(d))
}

// Now returns the current time of the clock.
func (c *MockClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.time
}

// NewTicker creates a new ticker that will fire once you advance the clock by the duration.
func (c *MockClock) NewTicker(d time.Duration) (<-chan time.Time, func()) {
	c.mu.Lock()
	defer c.mu.Unlock()

	ch := make(chan time.Time, 1)
	stopped := &atomic.Bool{}
	ticker := &mockTicker{nextTick: c.time, interval: d, ch: ch, stopped: stopped}
	c.tickers = append(c.tickers, ticker)
	stop := func() {
		stopped.Store(true)
	}

	return ch, stop
}

// NewTimer creates a new timer that will fire once you advance the clock by the duration.
func (c *MockClock) NewTimer(d time.Duration) (<-chan time.Time, func() bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	ch := make(chan time.Time, 1)
	stopped := &atomic.Bool{}

	// Fire the timer straight away if the duration is less than zero.
	if d <= 0 {
		ch <- c.time
		return ch, func() bool { return false }
	}

	timer := &mockTimer{deadline: c.time.Add(d), ch: ch, stopped: stopped}
	c.timers = append(c.timers, timer)
	stop := func() bool {
		return stopped.CompareAndSwap(false, true)
	}

	return ch, stop
}

// Since calculates the time since the given time t.
func (c *MockClock) Since(t time.Time) time.Duration {
	return c.Now().Sub(t)
}
