package truncate

import "time"

const millisecondDay int64 = 24 * 60 * 60 * 1000

// Day truncates the timestamp to the start of the day.
func Day(timestamp int64) int64 {
	return timestamp - (timestamp % millisecondDay)
}

// Week truncates the timestamp to the start of the week.
func Week(timestamp int64) int64 {
	t := time.UnixMilli(timestamp)
	for t.Weekday() != time.Monday {
		t = t.AddDate(0, 0, -1)
	}
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).UnixMilli()
}

// Month truncates the timestamp to the start of the month.
func Month(timestamp int64) int64 {
	t := time.UnixMilli(timestamp)
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location()).UnixMilli()
}

// Year truncates the timestamp to the start of the year.
func Year(timestamp int64) int64 {
	t := time.UnixMilli(timestamp)
	return time.Date(t.Year(), time.January, 1, 0, 0, 0, 0, t.Location()).UnixMilli()
}
