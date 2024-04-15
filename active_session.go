package pulse

import "log"

// CodingSession represents an ongoing coding session.
type CodingSession struct {
	// startStops is a slice of timestamps representing the start and stop times of
	// an ongoing coding session. This ensures accurate time tracking when switching
	// between multiple editor processes. We only count time for one at a time.
	startStops []int64
	bufStack   *bufferStack
	EditorID   string
	StartedAt  int64
	OS         string
	Editor     string
}

// StartSession creates a new coding session.
func StartSession(editorID string, startedAt int64, os, editor string) *CodingSession {
	return &CodingSession{
		startStops: []int64{startedAt},
		bufStack:   newBufferStack(),
		EditorID:   editorID,
		StartedAt:  startedAt,
		OS:         os,
		Editor:     editor,
	}
}

// Pause should be called when another editor process gains focus.
func (s *CodingSession) Pause(time int64) {
	if currentBuffer := s.bufStack.peek(); currentBuffer != nil {
		currentBuffer.Close(time)
	}
	s.startStops = append(s.startStops, time)
}

// PauseTime is used to determine which coding session to resume.
func (s *CodingSession) PauseTime() int64 {
	if len(s.startStops) == 0 {
		return 0
	}
	return s.startStops[len(s.startStops)-1]
}

// Resume should be called when the editor regains focus.
func (s *CodingSession) Resume(time int64) {
	if currentBuffer := s.bufStack.peek(); currentBuffer != nil {
		currentBuffer.Open(time)
	}
	s.startStops = append(s.startStops, time)
}

// PushBuffer pushes a new buffer to the current sessions buffer stack.
func (s *CodingSession) PushBuffer(buffer Buffer) {
	// Stop recording time for the previous buffer.
	if currentBuffer := s.bufStack.peek(); currentBuffer != nil {
		currentBuffer.Close(buffer.LastOpened())
	}
	s.bufStack.push(buffer)
}

// HasBuffers returns true if the coding session has opened any file backed buffers.
func (s *CodingSession) HasBuffers() bool {
	return len(s.bufStack.buffers) > 0
}

// Duration returns the total duration of the coding session.
func (s *CodingSession) Duration() int64 {
	var duration int64
	log.Println(s.startStops)
	for i := 0; i < len(s.startStops); i += 2 {
		duration += s.startStops[i+1] - s.startStops[i]
	}
	return duration
}

// Active returns true if the coding session is considered active.
func (s *CodingSession) Active() bool {
	return len(s.startStops)%2 == 1
}

// End ends the active coding sessions. It sets the total duration in
// milliseconds, and turns the stack of buffers into a slice of files.
func (s *CodingSession) End(endedAt int64) Session {
	if currentBuffer := s.bufStack.peek(); currentBuffer != nil && currentBuffer.IsOpen() {
		currentBuffer.Close(endedAt)
	}

	if s.Active() {
		s.startStops = append(s.startStops, endedAt)
	}

	return Session{
		StartedAt:  s.StartedAt,
		EndedAt:    endedAt,
		DurationMs: s.Duration(),
		OS:         s.OS,
		Editor:     s.Editor,
		Files:      s.bufStack.files(),
	}
}
