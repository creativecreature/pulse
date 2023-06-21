// Clock is a simple abstraction to allow for easy time based assertions
package clock

import "time"

type Clock struct{}

func New() Clock {
	return Clock{}
}

func (c Clock) GetTime() int64 {
	return time.Now().UTC().UnixMilli()
}
