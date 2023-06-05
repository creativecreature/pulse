package domain

// File represents a ActiveFile that has been opened in the editor
type ActiveFile struct {
	OpenedAt   int64
	ClosedAt   int64
	Name       string
	Repository string
	Path       string
	Filetype   string
	DurationMs int64
}

// NewActiveFile creates a new file
func NewActiveFile(name, repo, filetype, path string, openedAt int64) *ActiveFile {
	return &ActiveFile{
		Name:       name,
		Repository: repo,
		Filetype:   filetype,
		Path:       path,
		OpenedAt:   openedAt,
		ClosedAt:   0,
	}
}

// File represents the files for any given coding session
type File struct {
	Name       string `json:"name"`
	Path       string `json:"path"`
	Repository string `json:"repository"`
	Filetype   string `json:"filetype"`
	DurationMs int64  `json:"duration_ms"`
}

// DailyFile represents all the work that has been done in a patricular file for a
// given day
type DailyFile struct {
	Name       string `bson:"name"`
	Path       string `bson:"path"`
	Filetype   string `bson:"filetype"`
	DurationMs int64  `bson:"duration_ms"`
}
