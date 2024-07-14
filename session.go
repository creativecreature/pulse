package pulse

import (
	"cmp"
	"sort"
	"time"
)

// CodingSession represents a coding session that has been aggregated
// for a given time period (day, week, month, year).
type CodingSession struct {
	Date         time.Time     `json:"date"`
	Duration     time.Duration `json:"duration"`
	Repositories Repositories  `json:"repositories"`
}

// TruncateDay truncates the time to the start of the day.
func TruncateDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func NewCodingSession(buffers Buffers, now time.Time) CodingSession {
	repos := make(map[string]Repository)
	for _, buf := range buffers {
		repo, ok := repos[buf.Repository]
		if !ok {
			repo = Repository{Name: buf.Repository, Files: make(Files, 0)}
		}

		file := File{
			Name:     buf.Filename,
			Path:     buf.Filepath,
			Filetype: buf.Filetype,
			Duration: buf.Duration,
		}
		repo.Duration += file.Duration
		repo.Files = append(repo.Files, file)
		repos[buf.Repository] = repo
	}

	var totalDuration time.Duration
	repositories := make(Repositories, 0, len(repos))
	for _, repo := range repos {
		totalDuration += repo.Duration
		repositories = append(repositories, repo)
	}

	session := CodingSession{
		Date:         TruncateDay(now),
		Duration:     totalDuration,
		Repositories: repositories,
	}
	return session
}

// Merge takes two coding sessions, merges them, and returns the result.
func (a CodingSession) Merge(b CodingSession) CodingSession {
	mergedSession := CodingSession{
		Date:         cmp.Or(a.Date, b.Date),
		Duration:     a.Duration + b.Duration,
		Repositories: a.Repositories.merge(b.Repositories),
	}

	return mergedSession
}

// CodingSessions represents a slice of coding sessions.
type CodingSessions []CodingSession

func (s CodingSessions) Len() int {
	return len(s)
}

func (s CodingSessions) Less(i, j int) bool {
	return s[i].Date.Before(s[j].Date)
}

func (s CodingSessions) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func merge(sessions CodingSessions, truncate func(time.Time) time.Time) CodingSessions {
	truncatedDateAggregatedSession := make(map[time.Time]CodingSession)
	for _, s := range sessions {
		s.Date = truncate(s.Date)
		currentSession := truncatedDateAggregatedSession[s.Date]
		truncatedDateAggregatedSession[s.Date] = s.Merge(currentSession)
	}
	values := CodingSessions{}
	for _, v := range truncatedDateAggregatedSession {
		values = append(values, v)
	}
	sort.Sort(values)
	return values
}

// MergeByDay merges sessions that occurred the same day.
func (s CodingSessions) MergeByDay() CodingSessions {
	return merge(s, TruncateDay)
}

// MergeByWeek merges sessions that occurred the same week.
func (s CodingSessions) MergeByWeek() CodingSessions {
	truncate := func(t time.Time) time.Time {
		for t.Weekday() != time.Monday {
			t = t.AddDate(0, 0, -1)
		}
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	}
	return merge(s, truncate)
}

// MergeByMonth merges sessions that occurred the same month.
func (s CodingSessions) MergeByMonth() CodingSessions {
	truncate := func(t time.Time) time.Time {
		return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
	}
	return merge(s, truncate)
}

// MergeByYear merges sessions that occurred the same year.
func (s CodingSessions) MergeByYear() CodingSessions {
	truncate := func(t time.Time) time.Time {
		return time.Date(t.Year(), time.January, 1, 0, 0, 0, 0, t.Location())
	}
	return merge(s, truncate)
}
