package pulse

// AggregatedSession represents a session that has been aggregated
// for a given time period (day, week, month, year).
type AggregatedSession struct {
	ID           string       `bson:"_id,omitempty"`
	Period       TimePeriod   `bson:"period"`
	Date         int64        `bson:"date"`
	DateString   string       `bson:"date_string"`
	TotalTimeMs  int64        `bson:"total_time_ms"`
	Repositories Repositories `bson:"repositories"`
}

// merge takes two aggregated sessions, merges them, and returns the result.
func (a AggregatedSession) merge(b AggregatedSession, date int64, timePeriod TimePeriod) AggregatedSession {
	mergedSession := AggregatedSession{
		Period:       timePeriod,
		Date:         date,
		DateString:   a.DateString,
		TotalTimeMs:  a.TotalTimeMs + b.TotalTimeMs,
		Repositories: a.Repositories.merge(b.Repositories),
	}

	return mergedSession
}
