package domain

// ActiveSession represents a coding session that is currently ongoing within an editor
type ActiveSession struct {
	bufStack  *bufferStack
	StartedAt int64
	OS        string
	Editor    string
}

// StartSession creates a new active coding session
func StartSession(startedAt int64, os, editor string) *ActiveSession {
	return &ActiveSession{
		StartedAt: startedAt, OS: os,
		Editor:   editor,
		bufStack: &bufferStack{buffers: make([]Buffer, 0)},
	}
}

// PushBuffer pushes a new buffer to the sessions buffer stack
func (session *ActiveSession) PushBuffer(buffer Buffer) {
	// Stop recording time for the previous buffer (if we have one)
	currentBuffer := session.bufStack.peek()
	if currentBuffer != nil {
		currentBuffer.ClosedAt = buffer.OpenedAt
	}
	session.bufStack.push(buffer)
}

// files turns a slice of buffers into a slice of files
func files(buffers []Buffer) []File {
	files := make([]File, 0)
	for _, b := range buffers {
		files = append(files, fileFromBuffer(b))
	}
	return files
}

// End ends the active coding sessions and returns a Session struct which can be saved to disk
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
