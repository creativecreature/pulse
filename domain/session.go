package domain

import (
	"encoding/json"
)

// Session represents a "raw" coding session from an editor. They are stored in
// a temporary location, and are later aggregated by time period and moved to a
// permanent storage
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
