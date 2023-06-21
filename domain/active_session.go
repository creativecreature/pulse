package domain

// ActiveSession represents a coding session that is active in an editor
type ActiveSession struct {
	bufStack  *bufferStack
	StartedAt int64
	OS        string
	Editor    string
}

// StartSession creates a new active coding session
func StartSession(startedAt int64, os, editor string) *ActiveSession {
	return &ActiveSession{
		StartedAt: startedAt,
		OS:        os,
		Editor:    editor,
		bufStack:  &bufferStack{buffers: make([]Buffer, 0)},
	}
}

func (session *ActiveSession) PushBuffer(file Buffer) {
	// Stop recording time for the previous buffer if we have one
	currentBuffer := session.bufStack.peek()
	if currentBuffer != nil {
		currentBuffer.ClosedAt = file.OpenedAt
	}
	session.bufStack.push(file)
}

func files(buffers []Buffer) []File {
	files := make([]File, 0)
	for _, b := range buffers {
		files = append(files, fileFromBuffer(b))
	}
	return files
}

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
