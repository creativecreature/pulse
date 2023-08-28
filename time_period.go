package codeharvest

// Represents the time frame for which a sessions aggregation was performed.
type TimePeriod int8

const (
	Day TimePeriod = iota
	Week
	Month
	Year
)
