package domain

import (
	"errors"
)

// bufferStack represents the stack of buffers that have been opened during a
// coding session.
type bufferStack struct {
	s []Buffer
}

// push pushes a buffer onto the stack
func (s *bufferStack) push(f Buffer) {
	s.s = append(s.s, f)
}

func (s *bufferStack) peek() *Buffer {
	if len(s.s) == 0 {
		return nil
	}
	return &s.s[len(s.s)-1]
}

// pop pops a buffer off the stack
func (s *bufferStack) pop() (Buffer, error) {
	l := len(s.s)
	if l == 0 {
		return Buffer{}, errors.New("stack is empty")
	}

	res := s.s[l-1]
	s.s = s.s[:l-1]
	return res, nil
}

// list takes the stack of buffers, merges them, and returns a slice
func (s *bufferStack) list() []Buffer {
	mergedBuffers := map[string]Buffer{}
	for buffer, err := s.pop(); err == nil; buffer, err = s.pop() {
		if mergedBuffer, exists := mergedBuffers[buffer.Filepath]; !exists {
			buffer.DurationMs = buffer.ClosedAt - buffer.OpenedAt
			mergedBuffers[buffer.Filepath] = buffer
		} else {
			mergedBuffer.DurationMs += buffer.ClosedAt - buffer.OpenedAt
			mergedBuffers[buffer.Filepath] = mergedBuffer
		}
	}

	buffers := make([]Buffer, 0)
	for _, buffer := range mergedBuffers {
		buffers = append(buffers, buffer)
	}

	return buffers
}
