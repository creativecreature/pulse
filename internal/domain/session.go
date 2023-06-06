package domain

import (
	"encoding/json"
	"time"
)

// ActiveSession represents a coding session that is active in one of the clients.
type ActiveSession struct {
	Filestack       *filestack
	StartedAt       int64
	EndedAt         int64
	DurationMs      int64
	OS              string
	Editor          string
	AggregatedFiles map[string]*ActiveFile
}

// NewActiveSession creates a new active coding session
func NewActiveSession(startedAt int64, os, editor string) *ActiveSession {
	return &ActiveSession{
		StartedAt:       startedAt,
		OS:              os,
		Editor:          editor,
		Filestack:       &filestack{s: make([]*ActiveFile, 0)},
		AggregatedFiles: make(map[string]*ActiveFile),
	}
}

// Session represents one of many coding sessions that can occur during a
// given day. They are stored in a tmp directory, and later merged by a cron.
type Session struct {
	StartedAt  int64  `json:"started_at"`
	EndedAt    int64  `json:"ended_at"`
	DurationMs int64  `json:"duration_ms"`
	OS         string `json:"os"`
	Editor     string `json:"editor"`
	Files      []File `json:"files"`
}

func (session Session) Serialize() ([]byte, error) {
	return json.MarshalIndent(session, "", "  ")
}

func NewSession(s ActiveSession) Session {
	files := make([]File, 0)
	for _, f := range s.AggregatedFiles {
		file := File{
			Name:       f.Name,
			Path:       f.Path,
			Repository: f.Repository,
			Filetype:   f.Filetype,
			DurationMs: f.DurationMs,
		}
		files = append(files, file)
	}

	return Session{
		StartedAt:  s.StartedAt,
		EndedAt:    s.EndedAt,
		DurationMs: s.DurationMs,
		OS:         s.OS,
		Editor:     s.Editor,
		Files:      files,
	}
}

// Represents the time period that a session was aggregated for
type TimePeriod int8

const (
	Day TimePeriod = iota
	Week
	Month
	Year
)

// truncateDay is used to cluster unix timestamps into days
func truncateDay(timestamp int64) int64 {
	var dayInMs int64 = 24 * 60 * 60 * 1000
	return timestamp - (timestamp % dayInMs)
}

// AggregatedSession represents several TempSessions that have been merged together
// for a given interval.
type AggregatedSession struct {
	ID           string       `bson:"_id,omitempty"`
	Period       TimePeriod   `bson:"period"`
	Date         int64        `bson:"date"`
	DateString   string       `bson:"date_string"` // yyyy-mm-dd
	TotalTimeMs  int64        `bson:"total_time_ms"`
	Repositories []Repository `bson:"repositories"`
}

// groupByDay groups the temporary sessions by day
func groupByDay(session []Session) map[int64][]Session {
	buckets := make(map[int64][]Session)
	for _, s := range session {
		d := truncateDay(s.StartedAt)
		buckets[d] = append(buckets[d], s)
	}
	return buckets
}

// aggregateByDay takes a map where the key is the day and the value is a slice of
// temporary aggregateByDay that have occurred during that day. It returns the
// aggregated aggregateByDay.
func AggregateByDay(sessions []Session) []AggregatedSession {
	sessionsPerDay := groupByDay(sessions)
	aggregatedSessions := make([]AggregatedSession, 0)
	for date, tempSessions := range sessionsPerDay {
		dateString := time.Unix(0, date*int64(time.Millisecond)).Format("2006-01-02")
		var totalTime int64 = 0
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
