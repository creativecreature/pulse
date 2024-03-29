package codeharvest

import (
	"time"

	"github.com/creativecreature/code-harvest/truncate"
)

const yymmdd = "2006-01-02"

// Sessions is a slice of several Session structs.
type Sessions []Session

func (s Sessions) Len() int {
	return len(s)
}

func (s Sessions) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s Sessions) Less(i, j int) bool {
	return s[i].StartedAt < s[j].StartedAt
}

// groupByDay groups the sessions by day they occurred.
func groupByDay(session []Session) map[int64][]Session {
	buckets := make(map[int64][]Session)
	for _, s := range session {
		d := truncate.Day(s.StartedAt)
		buckets[d] = append(buckets[d], s)
	}
	return buckets
}

// Aggregate takes a slice of raw coding sessions and aggregates them by day.
func (s Sessions) Aggregate() []AggregatedSession {
	sessionsPerDay := groupByDay(s)
	aggregatedSessions := make([]AggregatedSession, 0)

	for date, tempSessions := range sessionsPerDay {
		dateString := time.Unix(0, date*int64(time.Millisecond)).Format(yymmdd)
		var totalTime int64
		for _, tempSession := range tempSessions {
			totalTime += tempSession.DurationMs
		}
		session := AggregatedSession{
			Period:       Day,
			Date:         date,
			DateString:   dateString,
			TotalTimeMs:  totalTime,
			Repositories: repositories(tempSessions),
		}
		aggregatedSessions = append(aggregatedSessions, session)
	}

	return aggregatedSessions
}
