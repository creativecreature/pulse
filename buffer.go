package pulse

import (
	"cmp"
	"fmt"
	"time"
)

// Buffer represents a buffer that has been opened during a coding session.
type Buffer struct {
	OpenedAt   time.Time     `json:"-"`
	ClosedAt   time.Time     `json:"-"`
	Duration   time.Duration `json:"duration"`
	Filename   string        `json:"filename"`
	Filepath   string        `json:"filepath"`
	Filetype   string        `json:"filetype"`
	Repository string        `json:"repository"`
}

// NewBuffer creates a new buffer.
func NewBuffer(filename, repo, filetype, filepath string, openedAt time.Time) Buffer {
	return Buffer{
		OpenedAt:   openedAt,
		Filename:   filename,
		Filepath:   filepath,
		Filetype:   filetype,
		Repository: repo,
	}
}

func (b *Buffer) Close(closedAt time.Time) {
	b.ClosedAt = closedAt
	b.Duration = b.ClosedAt.Sub(b.OpenedAt)
}

func (b *Buffer) Key() string {
	return fmt.Sprintf("%s_%s_%s", b.OpenedAt.Format("2006-01-02"), b.Repository, b.Filepath)
}

func (b *Buffer) Merge(other Buffer) Buffer {
	return Buffer{
		Filename:   cmp.Or(b.Filename, other.Filename),
		Filepath:   cmp.Or(b.Filepath, other.Filepath),
		Filetype:   cmp.Or(b.Filetype, other.Filetype),
		Repository: cmp.Or(b.Repository, other.Repository),
		Duration:   b.Duration + other.Duration,
	}
}

type Buffers []Buffer

func (b Buffers) Len() int {
	return len(b)
}

func (b Buffers) Less(i, j int) bool {
	return b[i].OpenedAt.Before(b[j].OpenedAt)
}

func (b Buffers) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}
