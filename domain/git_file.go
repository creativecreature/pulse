package domain

// GitFile represents a file within a git repository
type GitFile struct {
	Name       string
	Filetype   string
	Repository string
	Path       string
}
