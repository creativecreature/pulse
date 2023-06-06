package domain

// StoredFile represents a file that has been stored for a coding session
type StoredFile struct {
	Name       string `json:"name"`
	Path       string `json:"path"`
	Repository string `json:"repository"`
	Filetype   string `json:"filetype"`
	DurationMs int64  `json:"duration_ms"`
}
