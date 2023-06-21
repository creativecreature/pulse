package truncate

import "time"

const dayInMs int64 = 24 * 60 * 60 * 1000

func Day(timestamp int64) int64 {
	return timestamp - (timestamp % dayInMs)
}

func Week(timestamp int64) int64 {
	t := time.Unix(0, timestamp*int64(time.Millisecond))
	for t.Weekday() != time.Monday {
		t = t.AddDate(0, 0, -1)
	}
	t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	return t.UnixNano() / int64(time.Millisecond)
}

func Month(timestamp int64) int64 {
	t := time.Unix(0, timestamp*int64(time.Millisecond))
	t = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
	return t.UnixNano() / int64(time.Millisecond)
}

func Year(timestamp int64) int64 {
	t := time.Unix(0, timestamp*int64(time.Millisecond))
	t = time.Date(t.Year(), time.January, 1, 0, 0, 0, 0, t.Location())
	return t.UnixNano() / int64(time.Millisecond)
}
