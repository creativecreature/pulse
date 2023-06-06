package domain

import (
	"encoding/json"
)

// StoredSession represents one of many coding sessions that can occur during a given
// day. They are stored on disk and merged on a time interval by a cron job.
type StoredSession struct {
	StartedAt  int64        `json:"started_at"`
	EndedAt    int64        `json:"ended_at"`
	DurationMs int64        `json:"duration_ms"`
	OS         string       `json:"os"`
	Editor     string       `json:"editor"`
	Files      []StoredFile `json:"files"`
}

func (session StoredSession) Serialize() ([]byte, error) {
	return json.MarshalIndent(session, "", "  ")
}

func NewSession(s ActiveSession) StoredSession {
	files := make([]StoredFile, 0)

	for _, b := range s.MergedBuffers {
		file := StoredFile{
			Name:       b.Filename,
			Path:       b.Filepath,
			Repository: b.Repository,
			Filetype:   b.Filetype,
			DurationMs: b.DurationMs,
		}
		files = append(files, file)
	}

	return StoredSession{
		StartedAt:  s.StartedAt,
		EndedAt:    s.EndedAt,
		DurationMs: s.DurationMs,
		OS:         s.OS,
		Editor:     s.Editor,
		Files:      files,
	}
}
