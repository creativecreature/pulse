package clock

import "time"

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
