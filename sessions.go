package pulse

import (
	"time"

	"github.com/creativecreature/pulse/truncate"
)

// Sessions is a slice of several Session structs.
type Sessions []Session

func (s Sessions) Len() int {
	return len(s)
}

func (s Sessions) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s Sessions) Less(i, j int) bool {
	return s[i].StartedAt.Before(s[j].StartedAt)
}

// groupByDay groups a slice of sessions by day.
func groupByDay(session []Session) map[int64][]Session {
	buckets := make(map[int64][]Session)
	for _, s := range session {
		d := truncate.Day(s.StartedAt.UnixMilli())
		buckets[d] = append(buckets[d], s)
	}
	return buckets
}

// Aggregate takes a list of locally stored sessions and aggregates them by day.
func (s Sessions) Aggregate() AggregatedSessions {
	sessionsPerDay := groupByDay(s)
	aggregatedSessions := make(AggregatedSessions, 0)

	for date, tempSessions := range sessionsPerDay {
		var totalTimeMs int64
		for _, tempSession := range tempSessions {
			totalTimeMs += tempSession.Duration.Milliseconds()
		}
		session := AggregatedSession{
			Period:       Day,
			EpochDateMs:  date,
			DateString:   time.UnixMilli(date).Format("2006-01-02"),
			TotalTimeMs:  totalTimeMs,
			Repositories: repositories(tempSessions),
		}
		aggregatedSessions = append(aggregatedSessions, session)
	}

	return aggregatedSessions
}
