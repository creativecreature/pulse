package domain

// Session represents a coding session
type Session struct {
	Filestack       *filestack
	StartedAt       int64
	EndedAt         int64
	DurationMs      int64
	OS              string
	Editor          string
	AggregatedFiles map[string]*file
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
