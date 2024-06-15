package pulse

import "cmp"

// Repository represents a git repository. A coding session
// might open files across any number of repos. The files of
// the coding session are later grouped by repository.
type Repository struct {
	Name       string `bson:"name"`
	Files      Files  `bson:"files"`
	DurationMs int64  `bson:"duration_ms"`
}

// merge takes two repositories, merges them, and returns the result.
func (r Repository) merge(b Repository) Repository {
	return Repository{
		Name:       cmp.Or(r.Name, b.Name),
		Files:      r.Files.merge(b.Files),
		DurationMs: r.DurationMs + b.DurationMs,
	}
}
