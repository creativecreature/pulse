package pulse

import (
	"cmp"
	"time"
)

// File represents a file that has been opened during a coding session.
type File struct {
	Name       string        `json:"name"`
	Path       string        `json:"path"`
	Repository string        `json:"repository"`
	Filetype   string        `json:"filetype"`
	Duration   time.Duration `json:"duration"`
}

func (f File) merge(b File) File {
	return File{
		Name:       cmp.Or(f.Name, b.Name),
		Path:       cmp.Or(f.Path, b.Path),
		Repository: cmp.Or(f.Repository, b.Repository),
		Filetype:   cmp.Or(f.Filetype, b.Filetype),
		Duration:   f.Duration + b.Duration,
	}
}

// fileFromBuffer turns a code buffer into a file.
func fileFromBuffer(b Buffer) File {
	return File{
		Name:       b.Filename,
		Path:       b.Filepath,
		Repository: b.Repository,
		Filetype:   b.Filetype,
		Duration:   b.Duration(),
	}
}

// Files represents a list of files.
type Files []File
