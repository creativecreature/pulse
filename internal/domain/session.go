package domain

// Session represents a coding session
type Session struct {
	Filestack       *filestack       `bson:"-"`
	StartedAt       int64            `bson:"started_at"`
	EndedAt         int64            `bson:"ended_at"`
	DurationMs      int64            `bson:"duration_ms"`
	OS              string           `bson:"os"`
	Editor          string           `bson:"editor"`
	AggregatedFiles map[string]*file `bson:"files"`
}

// NewSession creates a new coding session
func NewSession(startedAt int64, os, editor string) *Session {
	return &Session{
		StartedAt:       startedAt,
		OS:              os,
		Editor:          editor,
		Filestack:       &filestack{s: make([]*file, 0)},
		AggregatedFiles: make(map[string]*file),
	}
}
