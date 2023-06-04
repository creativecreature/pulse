package models

import (
	"encoding/json"

	"code-harvest.conner.dev/internal/domain"
)

// NOTE: In this file, we have defined structs that correspond to the raw
// format of our coding sessions. These structs are used to save the data
// temporarily to disk. Subsequently, a cron job will aggregate and consolidate
// the data on a daily basis.

// TemporaryFile represents how we store coding session files on disk
type TemporaryFile struct {
	Name       string `json:"name"`
	Path       string `json:"path"`
	Repository string `json:"repository"`
	Filetype   string `json:"filetype"`
	DurationMs int64  `json:"duration_ms"`
}

// TemporarySession represents how we store coding session data on disk
type TemporarySession struct {
	StartedAt  int64           `json:"started_at"`
	EndedAt    int64           `json:"ended_at"`
	DurationMs int64           `json:"duration_ms"`
	OS         string          `json:"os"`
	Editor     string          `json:"editor"`
	Files      []TemporaryFile `json:"files"`
}

func NewTemporarySession(s domain.Session) TemporarySession {
	files := make([]TemporaryFile, 0)
	for _, f := range s.AggregatedFiles {
		file := TemporaryFile{
			Name:       f.Name,
			Path:       f.Path,
			Repository: f.Repository,
			Filetype:   f.Filetype,
			DurationMs: f.DurationMs,
		}
		files = append(files, file)
	}

	return TemporarySession{
		StartedAt:  s.StartedAt,
		EndedAt:    s.EndedAt,
		DurationMs: s.DurationMs,
		OS:         s.OS,
		Editor:     s.Editor,
		Files:      files,
	}
}

func (session TemporarySession) Serialize() ([]byte, error) {
	return json.MarshalIndent(session, "", "  ")
}
