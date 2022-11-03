package models

type file struct {
	OpenedAt   int64  `bson:"-"`
	ClosedAt   int64  `bson:"-"`
	Name       string `bson:"name"`
	Repository string `bson:"repository"`
	Path       string `bson:"path"`
	Filetype   string `bson:"filetype"`
	DurationMs int64  `bson:"duration_ms"`
}

func NewFile(name, repo, filetype, path string, openedAt int64) *file {
	return &file{
		Name:       name,
		Repository: repo,
		Filetype:   filetype,
		Path:       path,
		OpenedAt:   openedAt,
		ClosedAt:   0,
	}
}

type stack struct {
	s []*file
}

func (s *stack) Len() int {
	return len(s.s)
}

func (s *stack) Push(f *file) {
	s.s = append(s.s, f)
}

func (s *stack) Pop() *file {
	l := len(s.s)
	if l == 0 {
		return nil
	}

	res := s.s[l-1]
	s.s = s.s[:l-1]
	return res
}

func (s *stack) Peek() *file {
	l := len(s.s)
	if l == 0 {
		return nil
	}
	return s.s[len(s.s)-1]
}

func NewSession(startedAt int64, os, editor string) *Session {
	return &Session{
		StartedAt:       startedAt,
		OS:              os,
		Editor:          editor,
		FileStack:       &stack{s: make([]*file, 0)},
		AggregatedFiles: make(map[string]*file),
	}
}

type Session struct {
	FileStack       *stack           `bson:"-"`
	StartedAt       int64            `bson:"started_at"`
	EndedAt         int64            `bson:"ended_at"`
	DurationMs      int64            `bson:"duration_ms"`
	OS              string           `bson:"os"`
	Editor          string           `bson:"editor"`
	AggregatedFiles map[string]*file `bson:"files"`
}
