package domain

// ActiveSession represents a coding session that is active in an editor
type ActiveSession struct {
	bufStack   *bufferStack
	StartedAt  int64
	EndedAt    int64
	DurationMs int64
	OS         string
	Editor     string
}

// NewActiveSession creates a new active coding session
func NewActiveSession(startedAt int64, os, editor string) *ActiveSession {
	return &ActiveSession{
		StartedAt: startedAt,
		OS:        os,
		Editor:    editor,
		bufStack:  &bufferStack{s: make([]Buffer, 0)},
	}
}

func (session *ActiveSession) closeCurrentBuffer(closedAt int64) {
	currentBuffer := session.bufStack.peek()
	if currentBuffer != nil {
		currentBuffer.ClosedAt = closedAt
	}
}

func (session *ActiveSession) PushBuffer(file Buffer) {
	session.closeCurrentBuffer(file.OpenedAt)
	session.bufStack.push(file)
}

func (s *ActiveSession) End(time int64) Session {
	s.closeCurrentBuffer(time)
	s.DurationMs = time - s.StartedAt
	buffers := s.bufStack.list()

	files := make([]File, 0)
	for _, b := range buffers {
		file := File{
			Name:       b.Filename,
			Path:       b.Filepath,
			Repository: b.Repository,
			Filetype:   b.Filetype,
			DurationMs: b.DurationMs,
		}
		files = append(files, file)
	}

	return Session{
		StartedAt:  s.StartedAt,
		EndedAt:    time,
		DurationMs: s.DurationMs,
		OS:         s.OS,
		Editor:     s.Editor,
		Files:      files,
	}
}
