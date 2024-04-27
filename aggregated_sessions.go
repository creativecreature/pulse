package pulse

import (
	"github.com/creativecreature/pulse/truncate"
	"golang.org/x/exp/maps"
)

type AggregatedSessions []AggregatedSession

func merge(sessions AggregatedSessions, truncate func(int64) int64, timePeriod Period) AggregatedSessions {
	truncatedDateAggregatedSession := make(map[int64]AggregatedSession)
	for _, s := range sessions {
		truncatedDate := truncate(s.EpochDateMs)
		currentSession := truncatedDateAggregatedSession[truncatedDate]
		truncatedDateAggregatedSession[truncatedDate] = s.merge(currentSession, truncatedDate, timePeriod)
	}
	return maps.Values(truncatedDateAggregatedSession)
}

// MergeByDay merges sessions that occurred the same day.
func (s AggregatedSessions) MergeByDay() AggregatedSessions {
	return merge(s, truncate.Day, Day)
}

// MergeByWeek merges sessions that occurred the same week.
func (s AggregatedSessions) MergeByWeek() AggregatedSessions {
	return merge(s, truncate.Week, Week)
}

// MergeByWeek merges sessions that occurred the same month.
func (s AggregatedSessions) MergeByMonth() AggregatedSessions {
	return merge(s, truncate.Month, Month)
}

// MergeByYear merges sessions that occurred the same year.
func (s AggregatedSessions) MergeByYear() AggregatedSessions {
	return merge(s, truncate.Year, Year)
}
