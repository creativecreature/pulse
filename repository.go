package codeharvest

// Repository represents a git repository. A coding session
// might open files across any number of repos. The files of
// the coding session are later grouped by repository.
type Repository struct {
	Name       string          `bson:"name"`
	Files      AggregatedFiles `bson:"files"`
	DurationMs int64           `bson:"duration_ms"`
}

// merge takes two repositories, merges them, and returns the result.
func (a Repository) merge(b Repository) Repository {
	return Repository{
		Name:       a.Name,
		Files:      a.Files.merge(b.Files),
		DurationMs: a.DurationMs + b.DurationMs,
	}
}

// repositoryPathFile takes a slice of coding sessions and generates
// a map. In this map, the key is the repository name, and the value
// is another map wherein the key is the relative file path within the
// repository, and the value is the file itself. The files are merged
// if any of the sessions has worked on the same file.
func repositoryPathFile(sessions []Session) map[string]map[string]AggregatedFile {
	repoPathFile := make(map[string]map[string]AggregatedFile)
	for _, session := range sessions {
		for _, file := range session.Files {
			// Create an aggregated file from the file.
			aggregatedFile := AggregatedFile{
				Name:       file.Name,
				Path:       file.Path,
				Filetype:   file.Filetype,
				DurationMs: file.DurationMs,
			}

			// Check if it is the first time we're seeing a repository
			// with this name. In that case, we'll initialize the map.
			if _, ok := repoPathFile[file.Repository]; !ok {
				repoPathFile[file.Repository] = make(map[string]AggregatedFile)
			}

			// Check if it is the first time we're seeing this file in this
			// repository. If it is not the first time, we'll merge them.
			if f, ok := repoPathFile[file.Repository][file.Path]; !ok {
				repoPathFile[file.Repository][file.Path] = aggregatedFile
			} else {
				repoPathFile[file.Repository][file.Path] = f.merge(aggregatedFile)
			}
		}
	}
	return repoPathFile
}
