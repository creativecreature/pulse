package domain

import "encoding/json"

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
