package pulse

import "time"

const millisecondDay int64 = 24 * 60 * 60 * 1000

// TruncateDay truncates the timestamp to the start of the day.
func TruncateDay(timestamp int64) int64 {
	return timestamp - (timestamp % millisecondDay)
}

// TruncateWeek truncates the timestamp to the start of the week.
func TruncateWeek(timestamp int64) int64 {
	t := time.UnixMilli(timestamp)
	for t.Weekday() != time.Monday {
		t = t.AddDate(0, 0, -1)
	}
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).UnixMilli()
}

// TruncateMonth truncates the timestamp to the start of the month.
func TruncateMonth(timestamp int64) int64 {
	t := time.UnixMilli(timestamp)
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location()).UnixMilli()
}

// TruncateYear truncates the timestamp to the start of the year.
func TruncateYear(timestamp int64) int64 {
	t := time.UnixMilli(timestamp)
	return time.Date(t.Year(), time.January, 1, 0, 0, 0, 0, t.Location()).UnixMilli()
}
