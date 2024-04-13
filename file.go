package pulse

// File represents a file that has been opened during a coding session.
type File struct {
	Name       string `json:"name"`
	Path       string `json:"path"`
	Repository string `json:"repository"`
	Filetype   string `json:"filetype"`
	DurationMs int64  `json:"duration_ms"`
}

// fileFromBuffer turns a code buffer into a file.
func fileFromBuffer(b Buffer) File {
	return File{
		Name:       b.Filename,
		Path:       b.Filepath,
		Repository: b.Repository,
		Filetype:   b.Filetype,
		DurationMs: b.Duration(),
	}
}

// Files represents a list of files.
type Files []File
