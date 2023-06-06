package domain

// bufferStack represents the stack of buffers in an active coding session
type bufferStack struct {
	s []*Buffer
}

// push pushes a buffer onto the stack
func (s *bufferStack) push(f *Buffer) {
	s.s = append(s.s, f)
}

// pop pops a file off the stack
func (s *bufferStack) pop() *Buffer {
	l := len(s.s)
	if l == 0 {
		return nil
	}

	res := s.s[l-1]
	s.s = s.s[:l-1]
	return res
}

// peek returns a pointer to the most recently opened buffer
func (s *bufferStack) peek() *Buffer {
	l := len(s.s)
	if l == 0 {
		return nil
	}
	return s.s[len(s.s)-1]
}
