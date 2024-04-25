package pulse

import "time"

// Buffer represents a buffer that has been opened during an active coding session.
type Buffer struct {
	OpenedClosed []time.Time
	Filename     string
	Repository   string
	Filepath     string
	Filetype     string
}

// NewBuffer creates a new buffer.
func NewBuffer(filename, repo, filetype, filepath string, openedAt time.Time) Buffer {
	return Buffer{
		OpenedClosed: []time.Time{openedAt},
		Filename:     filename,
		Repository:   repo,
		Filetype:     filetype,
		Filepath:     filepath,
	}
}

// Open should be called when a buffer is opened.
func (b *Buffer) Open(time time.Time) {
	b.OpenedClosed = append(b.OpenedClosed, time)
}

// IsOpen returns true if the given buffer is currently open.
func (b *Buffer) IsOpen() bool {
	return len(b.OpenedClosed)%2 == 1
}

// LastOpened returns the last time the buffer was opened.
func (b *Buffer) LastOpened() time.Time {
	return b.OpenedClosed[len(b.OpenedClosed)-1]
}

// Close should be called when a buffer is closed.
func (b *Buffer) Close(time time.Time) {
	b.OpenedClosed = append(b.OpenedClosed, time)
}

// Duration returns the total duration that the buffer has been open.
func (b *Buffer) Duration() time.Duration {
	var duration time.Duration
	for i := 0; i < len(b.OpenedClosed); i += 2 {
		duration += b.OpenedClosed[i+1].Sub(b.OpenedClosed[i])
	}
	return duration
}
