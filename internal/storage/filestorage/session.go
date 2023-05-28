package filestorage

import (
	"encoding/json"

	"code-harvest.conner.dev/internal/domain"
)

// File represents a file during the coding session
type File struct {
	Name       string `json:"name"`
	Repository string `json:"repository"`
	Path       string `json:"path"`
	Filetype   string `json:"filetype"`
	DurationMs int64  `json:"duration_ms"`
}

// Session represents the actual coding session
type Session struct {
	StartedAt  int64  `json:"started_at"`
	EndedAt    int64  `json:"ended_at"`
	DurationMs int64  `json:"duration_ms"`
	OS         string `json:"os"`
	Editor     string `json:"editor"`
	Files      []File `json:"files"`
}

func newSession(s domain.Session) Session {
	files := make([]File, 0)
	for _, f := range s.AggregatedFiles {
		file := File{
			Name:       f.Name,
			Repository: f.Repository,
			Path:       f.Path,
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

func (s Session) serialize() ([]byte, error) {
	return json.MarshalIndent(s, "", "  ")
}
