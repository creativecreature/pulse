package mock

type Clock struct {
	time int64
}

func (c *Clock) GetTime() int64 {
	return c.time
}

func (c *Clock) SetTime(time int64) {
	c.time = time
}
