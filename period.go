package pulse

// Period represents the time period for which the coding sessions have been aggregated.
type Period int8

const (
	Day Period = iota
	Week
	Month
	Year
)
