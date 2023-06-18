package domain

import (
	"encoding/json"
)

// Session represents one of many coding sessions that can occur during a given
// day. They are stored in a temporary location, and are later aggregated by
// time period and moved to a permanent storage location
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

// NewSession is used to create a Session from an ActiveSession. The Session
// can be saved to disk and aggregated by time period.
func NewSession(s ActiveSession) Session {
	files := make([]File, 0)

	for _, b := range s.MergedBuffers {
		file := File{
			Name:       b.Filename,
			Path:       b.Filepath,
			Repository: b.Repository,
			Filetype:   b.Filetype,
			DurationMs: b.DurationMs,
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
