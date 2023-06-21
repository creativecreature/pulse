package domain

import (
	"code-harvest.conner.dev/truncate"
	"golang.org/x/exp/maps"
)

type AggregatedSessions []AggregatedSession

func merge(sessions AggregatedSessions, createKey func(s AggregatedSession) int64, timePeriod TimePeriod) AggregatedSessions {
	sessionMap := make(map[int64]AggregatedSession)
	for _, s := range sessions {
		key := createKey(s)
		if session, ok := sessionMap[key]; !ok {
			sessionMap[key] = s
		} else {
			sessionMap[key] = s.merge(session, createKey(s), timePeriod)
		}
	}
	return maps.Values(sessionMap)
}

func (sessions AggregatedSessions) MergeByDay() AggregatedSessions {
	keyFunc := func(s AggregatedSession) int64 {
		return truncate.Day(s.Date)
	}
	return merge(sessions, keyFunc, Day)
}

func (sessions AggregatedSessions) MergeByWeek() AggregatedSessions {
	keyFunc := func(s AggregatedSession) int64 {
		return truncate.Week(s.Date)
	}
	return merge(sessions, keyFunc, Week)
}

func (sessions AggregatedSessions) MergeByMonth() AggregatedSessions {
	keyFunc := func(s AggregatedSession) int64 {
		return truncate.Month(s.Date)
	}
	return merge(sessions, keyFunc, Month)
}

func (sessions AggregatedSessions) MergeByYear() AggregatedSessions {
	keyFunc := func(s AggregatedSession) int64 {
		return truncate.Year(s.Date)
	}
	return merge(sessions, keyFunc, Year)
}
