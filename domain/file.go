package domain

// File represents a file that has been stored for a coding session
type File struct {
	Name       string `json:"name"`
	Path       string `json:"path"`
	Repository string `json:"repository"`
	Filetype   string `json:"filetype"`
	DurationMs int64  `json:"duration_ms"`
}

func fileFromBuffer(b Buffer) File {
	return File{
		Name:       b.Filename,
		Path:       b.Filepath,
		Repository: b.Repository,
		Filetype:   b.Filetype,
		DurationMs: b.DurationMs,
	}
}
