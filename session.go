package pulse

import (
	"encoding/json"
)

// Session is the raw representation of a past coding session. These sessions are
// stored temporarily on disk, and are later aggregated and and moved to a database.
type Session struct {
	StartedAt  int64  `json:"started_at"`
	EndedAt    int64  `json:"ended_at"`
	DurationMs int64  `json:"duration_ms"`
	OS         string `json:"os"`
	Editor     string `json:"editor"`
	Files      Files  `json:"files"`
}

// Serialize serializes the session to a JSON byte slice.
func (s Session) Serialize() ([]byte, error) {
	return json.MarshalIndent(s, "", "  ")
}
