package pulse

import (
	"encoding/json"
	"time"
)

// Session is the raw representation of a past coding session. These sessions are
// stored temporarily on disk, and are later aggregated and and moved to a database.
type Session struct {
	StartedAt time.Time     `json:"started_at"`
	EndedAt   time.Time     `json:"ended_at"`
	Duration  time.Duration `json:"duration"`
	OS        string        `json:"os"`
	Editor    string        `json:"editor"`
	Files     Files         `json:"files"`
}

// Serialize serializes the session to a JSON byte slice.
func (s Session) Serialize() ([]byte, error) {
	return json.MarshalIndent(s, "", "  ")
}
