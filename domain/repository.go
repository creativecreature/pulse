package domain

// Repository represents all work that has been done in a repository during a day
type Repository struct {
	Name       string          `bson:"name"`
	Files      AggregatedFiles `bson:"files"`
	DurationMs int64           `bson:"duration_ms"`
}

func (a Repository) merge(b Repository) Repository {
	return Repository{
		Name:       a.Name,
		Files:      a.Files.merge(b.Files),
		DurationMs: a.DurationMs + b.DurationMs,
	}
}

func aggregateFilesByRepo(sessions []Session) map[string]map[string]AggregatedFile {
	repositoryFiles := make(map[string]map[string]AggregatedFile)
	for _, session := range sessions {
		for _, file := range session.Files {
			// Create an aggregated file from the file
			aggregatedFile := AggregatedFile{
				Name:       file.Name,
				Path:       file.Path,
				Filetype:   file.Filetype,
				DurationMs: file.DurationMs,
			}

			// Check if it is the first time we're seeing this repo. In that case
			// we'll initialize the map.
			if _, ok := repositoryFiles[file.Repository]; !ok {
				repositoryFiles[file.Repository] = make(map[string]AggregatedFile)
			}

			// Check if it's the first time we're seeing this file in this repo.
			// If it's not the first time we'll merge them.
			if f, ok := repositoryFiles[file.Repository][file.Path]; !ok {
				repositoryFiles[file.Repository][file.Path] = aggregatedFile
			} else {
				repositoryFiles[file.Repository][file.Path] = f.merge(aggregatedFile)
			}
		}
	}
	return repositoryFiles
}
