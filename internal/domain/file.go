package domain

// File represents a file in the editor
type file struct {
	OpenedAt   int64  `bson:"-"`
	ClosedAt   int64  `bson:"-"`
	Name       string `bson:"name"`
	Repository string `bson:"repository"`
	Path       string `bson:"path"`
	Filetype   string `bson:"filetype"`
	DurationMs int64  `bson:"duration_ms"`
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

// Filestack is a stack of files that has been opened during a coding session
type filestack struct {
	s []*file
}

func (s *filestack) Len() int {
	return len(s.s)
}

func (s *filestack) Push(f *file) {
	s.s = append(s.s, f)
}

func (s *filestack) Pop() *file {
	l := len(s.s)
	if l == 0 {
		return nil
	}

	res := s.s[l-1]
	s.s = s.s[:l-1]
	return res
}

// Peek returns a pointer to the last file that was opened
func (s *filestack) Peek() *file {
	l := len(s.s)
	if l == 0 {
		return nil
	}
	return s.s[len(s.s)-1]
}
