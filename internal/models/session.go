package models

type Session struct {
	CurrentFile *File            `bson:"-"`
	OpenFiles   []*File          `bson:"-"`
	StartedAt   int64            `bson:"started_at"`
	EndedAt     int64            `bson:"ended_at"`
	DurationMs  int64            `bson:"duration_ms"`
	OS          string           `bson:"os"`
	Editor      string           `bson:"editor"`
	Files       map[string]*File `bson:"files"`
}
