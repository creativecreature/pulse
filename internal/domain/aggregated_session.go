package domain

// AggregatedSession represents several TempSessions that have been merged together
// for a given interval.
type AggregatedSession struct {
	ID           string       `bson:"_id,omitempty"`
	Period       TimePeriod   `bson:"period"`
	Date         int64        `bson:"date"`
	DateString   string       `bson:"date_string"` // yyyy-mm-dd
	TotalTimeMs  int64        `bson:"total_time_ms"`
	Repositories Repositories `bson:"repositories"`
}
