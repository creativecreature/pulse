package pulse

// Period represents the time period for which the data has been aggregated.
type Period int8

const (
	Day Period = iota
	Week
	Month
	Year
)
