package codeharvest

import (
	"github.com/creativecreature/code-harvest/truncate"
	"golang.org/x/exp/maps"
)

type AggregatedSessions []AggregatedSession

type truncateTimeFunc func(AggregatedSession) int64

// merge takes a slice of aggregated sessions as well as a truncate function,
// which is used to cluster and merge the sessions by time period.
func merge(sessions AggregatedSessions, truncate truncateTimeFunc, timePeriod TimePeriod) AggregatedSessions {
	sessionMap := make(map[int64]AggregatedSession)
	for _, s := range sessions {
		key := truncate(s)
		if session, ok := sessionMap[key]; !ok {
			sessionMap[key] = s
		} else {
			sessionMap[key] = s.merge(session, truncate(s), timePeriod)
		}
	}

	return maps.Values(sessionMap)
}

// MergeByDay merges sessions that occurred the same day.
func (s AggregatedSessions) MergeByDay() AggregatedSessions {
	truncateFunc := func(s AggregatedSession) int64 {
		return truncate.Day(s.Date)
	}
	return merge(s, truncateFunc, Day)
}

// MergeByWeek merges sessions that occurred the same week.
func (s AggregatedSessions) MergeByWeek() AggregatedSessions {
	truncateFunc := func(s AggregatedSession) int64 {
		return truncate.Week(s.Date)
	}
	return merge(s, truncateFunc, Week)
}

// MergeByWeek merges sessions that occurred the same month.
func (s AggregatedSessions) MergeByMonth() AggregatedSessions {
	truncateFunc := func(s AggregatedSession) int64 {
		return truncate.Month(s.Date)
	}
	return merge(s, truncateFunc, Month)
}

// MergeByYear merges sessions that occurred the same year.
func (s AggregatedSessions) MergeByYear() AggregatedSessions {
	truncateFunc := func(s AggregatedSession) int64 {
		return truncate.Year(s.Date)
	}
	return merge(s, truncateFunc, Year)
}
