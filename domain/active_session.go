package domain

// ActiveSession represents an ongoing coding session.
type ActiveSession struct {
	bufStack  *bufferStack
	StartedAt int64
	OS        string
	Editor    string
}

// StartSession creates a new active coding session.
func StartSession(startedAt int64, os, editor string) *ActiveSession {
	return &ActiveSession{
		StartedAt: startedAt, OS: os,
		Editor:   editor,
		bufStack: &bufferStack{buffers: make([]Buffer, 0)},
	}
}

// PushBuffer pushes a new buffer to the current sessions buffer stack.
func (session *ActiveSession) PushBuffer(buffer Buffer) {
	// Stop recording time for the previous buffer.
	if currentBuffer := session.bufStack.peek(); currentBuffer != nil {
		currentBuffer.ClosedAt = buffer.OpenedAt
	}
	session.bufStack.push(buffer)
}

// files turns a slice of buffers into a slice of files.
func files(buffers []Buffer) []File {
	files := make([]File, 0)
	for _, b := range buffers {
		files = append(files, fileFromBuffer(b))
	}
	return files
}

// End ends the active coding sessions. It sets the total duration in
// milliseconds, and turns the stack of buffers into a slice of files.
func (session *ActiveSession) End(endedAt int64) Session {
	currentBuffer := session.bufStack.peek()
	if currentBuffer != nil {
		currentBuffer.ClosedAt = endedAt
	}

	return Session{
		StartedAt:  session.StartedAt,
		EndedAt:    endedAt,
		DurationMs: endedAt - session.StartedAt,
		OS:         session.OS,
		Editor:     session.Editor,
		Files:      files(session.bufStack.slice()),
	}
}
