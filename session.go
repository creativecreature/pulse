package pulse

import (
	"cmp"
	"time"

	"golang.org/x/exp/maps"
)

// Period represents the time period for which the coding sessions have been aggregated.
type Period int8

const (
	Day Period = iota
	Week
	Month
	Year
)

// CodingSession represents a coding session that has been aggregated
// for a given time period (day, week, month, year).
type CodingSession struct {
	ID           string       `bson:"_id,omitempty"`
	Period       Period       `bson:"period"`
	EpochDateMs  int64        `bson:"date"`
	DateString   string       `bson:"date_string"`
	TotalTimeMs  int64        `bson:"total_time_ms"`
	Repositories Repositories `bson:"repositories"`
}

func NewCodingSession(buffers Buffers, now time.Time) CodingSession {
	repos := make(map[string]Repository)
	for _, buf := range buffers {
		repo, ok := repos[buf.Repository]
		if !ok {
			repo = Repository{Name: buf.Repository, Files: make(Files, 0)}
		}

		file := File{
			Name:       buf.Filename,
			Path:       buf.Filepath,
			Filetype:   buf.Filetype,
			DurationMs: buf.Duration.Milliseconds(),
		}
		repo.DurationMs += file.DurationMs
		repo.Files = append(repo.Files, file)
		repos[buf.Repository] = repo
	}

	var totalDurationMS int64
	repositories := make(Repositories, 0, len(repos))
	for _, repo := range repos {
		totalDurationMS += repo.DurationMs
		repositories = append(repositories, repo)
	}

	session := CodingSession{
		Period:       Day,
		EpochDateMs:  TruncateDay(now.UnixMilli()),
		DateString:   now.Format("2006-01-02"),
		TotalTimeMs:  totalDurationMS,
		Repositories: repositories,
	}
	return session
}

// merge takes two coding sessions, merges them, and returns the result.
func (a CodingSession) merge(b CodingSession, epochDateMs int64, timePeriod Period) CodingSession {
	mergedSession := CodingSession{
		Period:       timePeriod,
		EpochDateMs:  epochDateMs,
		DateString:   cmp.Or(a.DateString, b.DateString),
		TotalTimeMs:  a.TotalTimeMs + b.TotalTimeMs,
		Repositories: a.Repositories.merge(b.Repositories),
	}

	return mergedSession
}

// CodingSessions represents a slice of coding sessions.
type CodingSessions []CodingSession

// merge takes two slices of coding sessions, merges them, and returns the result.
func merge(sessions CodingSessions, truncate func(int64) int64, timePeriod Period) CodingSessions {
	truncatedDateAggregatedSession := make(map[int64]CodingSession)
	for _, s := range sessions {
		truncatedDate := truncate(s.EpochDateMs)
		currentSession := truncatedDateAggregatedSession[truncatedDate]
		truncatedDateAggregatedSession[truncatedDate] = s.merge(currentSession, truncatedDate, timePeriod)
	}
	return maps.Values(truncatedDateAggregatedSession)
}

// MergeByDay merges sessions that occurred the same day.
func (s CodingSessions) MergeByDay() CodingSessions {
	return merge(s, TruncateDay, Day)
}

// MergeByWeek merges sessions that occurred the same week.
func (s CodingSessions) MergeByWeek() CodingSessions {
	return merge(s, TruncateWeek, Week)
}

// MergeByWeek merges sessions that occurred the same month.
func (s CodingSessions) MergeByMonth() CodingSessions {
	return merge(s, TruncateMonth, Month)
}

// MergeByYear merges sessions that occurred the same year.
func (s CodingSessions) MergeByYear() CodingSessions {
	return merge(s, TruncateYear, Year)
}
