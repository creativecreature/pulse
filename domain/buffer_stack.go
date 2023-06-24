package domain

import (
	"errors"

	"golang.org/x/exp/maps"
)

// bufferStack is the stack of buffers that we have opened during an active coding session
type bufferStack struct {
	buffers []Buffer
}

// push pushes a buffer onto the stack
func (s *bufferStack) push(f Buffer) {
	s.buffers = append(s.buffers, f)
}

// peek returns a pointer to the most recent buffer
func (s *bufferStack) peek() *Buffer {
	if len(s.buffers) == 0 {
		return nil
	}
	return &s.buffers[len(s.buffers)-1]
}

// pop pops a buffer off the stack
func (s *bufferStack) pop() (Buffer, error) {
	length := len(s.buffers)
	if length == 0 {
		return Buffer{}, errors.New("stack is empty")
	}

	res := s.buffers[length-1]
	s.buffers = s.buffers[:length-1]
	return res, nil
}

// slice takes the stack of buffers, merges them by filepath, and returns the result
func (s *bufferStack) slice() []Buffer {
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
	return maps.Values(mergedBuffers)
}
