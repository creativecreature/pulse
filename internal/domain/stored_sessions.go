package domain

import "time"

const yymmdd = "2006-01-02"

type StoredSessions []StoredSession

const DayInMs int64 = 24 * 60 * 60 * 1000

func truncateDay(timestamp int64) int64 {
	return timestamp - (timestamp % DayInMs)
}

// groupByDay groups the temporary sessions by day
func groupByDay(session []StoredSession) map[int64][]StoredSession {
	buckets := make(map[int64][]StoredSession)
	for _, s := range session {
		d := truncateDay(s.StartedAt)
		buckets[d] = append(buckets[d], s)
	}
	return buckets
}

func (sessions StoredSessions) AggregateByDay() []AggregatedSession {
	sessionsPerDay := groupByDay(sessions)
	aggregatedSessions := make([]AggregatedSession, 0)

	for date, tempSessions := range sessionsPerDay {
		dateString := time.Unix(0, date*int64(time.Millisecond)).Format(yymmdd)
		var totalTime int64 = 0
		for _, tempSession := range tempSessions {
			totalTime += tempSession.DurationMs
		}
		session := AggregatedSession{
			Period:       Day,
			Date:         date,
			DateString:   dateString,
			TotalTimeMs:  totalTime,
			Repositories: repositoriesFromSessions(tempSessions),
		}
		aggregatedSessions = append(aggregatedSessions, session)
	}

	return aggregatedSessions
}
