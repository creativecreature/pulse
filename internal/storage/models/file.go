package models

// TemporaryFile represents how we store coding session files on disk
type TemporaryFile struct {
	Name       string `json:"name"`
	Repository string `json:"repository"`
	Filetype   string `json:"filetype"`
	DurationMs int64  `json:"duration_ms"`
}
