package domain

// ActiveSession represents a coding session that is active in one of the clients.
type ActiveSession struct {
	bufStack      *bufferStack
	StartedAt     int64
	EndedAt       int64
	DurationMs    int64
	OS            string
	Editor        string
	MergedBuffers map[string]*Buffer
}

// NewActiveSession creates a new active coding session
func NewActiveSession(startedAt int64, os, editor string) *ActiveSession {
	return &ActiveSession{
		StartedAt:     startedAt,
		OS:            os,
		Editor:        editor,
		bufStack:      &bufferStack{s: make([]*Buffer, 0)},
		MergedBuffers: make(map[string]*Buffer),
	}
}

func (session *ActiveSession) PeekBuffer() *Buffer {
	return session.bufStack.peek()
}

func (session *ActiveSession) PushBuffer(file *Buffer) {
	session.bufStack.push(file)
}

func (session *ActiveSession) PopBuffer() *Buffer {
	return session.bufStack.pop()
}
