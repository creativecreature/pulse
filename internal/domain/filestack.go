package domain

// filestack represents the stack of file that has been opened during a coding session
type filestack struct {
	s []*ActiveFile
}

func (s *filestack) Len() int {
	return len(s.s)
}

// Push pushes a file onto the stack
func (s *filestack) Push(f *ActiveFile) {
	s.s = append(s.s, f)
}

// Pop pops a file off the stack
func (s *filestack) Pop() *ActiveFile {
	l := len(s.s)
	if l == 0 {
		return nil
	}

	res := s.s[l-1]
	s.s = s.s[:l-1]
	return res
}

// Peek returns a pointer to the most recently opened file
func (s *filestack) Peek() *ActiveFile {
	l := len(s.s)
	if l == 0 {
		return nil
	}
	return s.s[len(s.s)-1]
}
