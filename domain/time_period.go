package domain

// Represents the time period that a session was aggregated for
type TimePeriod int8

const (
	Day TimePeriod = iota
	Week
	Month
	Year
)
