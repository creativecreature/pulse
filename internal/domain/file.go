package domain

// File represents a file that has been opened in the editor
type file struct {
	OpenedAt   int64
	ClosedAt   int64
	Name       string
	Repository string
	Path       string
	Filetype   string
	DurationMs int64
}

// NewFile creates a new file
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

// filestack represents the stack of file that has been opened during a coding session
type filestack struct {
	s []*file
}

func (s *filestack) Len() int {
	return len(s.s)
}

// Push pushes a file onto the stack
func (s *filestack) Push(f *file) {
	s.s = append(s.s, f)
}

// Pop pops a file off the stack
func (s *filestack) Pop() *file {
	l := len(s.s)
	if l == 0 {
		return nil
	}

	res := s.s[l-1]
	s.s = s.s[:l-1]
	return res
}

// Peek returns a pointer to the most recently opened file
func (s *filestack) Peek() *file {
	l := len(s.s)
	if l == 0 {
		return nil
	}
	return s.s[len(s.s)-1]
}
