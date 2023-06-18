package domain

// Repository represents all work that has been done in a repository during a day
type Repository struct {
	Name       string          `bson:"name"`
	Files      AggregatedFiles `bson:"files"`
	DurationMs int64           `bson:"duration_ms"`
}

func repositoryFileMap(sessions []Session) map[string]map[string]*AggregatedFile {
	repositoryFiles := make(map[string]map[string]*AggregatedFile)
	for _, session := range sessions {
		for _, file := range session.Files {
			// Check if it is the first time we're seeing this repo. In that case
			// we'll initialize the map.
			if _, ok := repositoryFiles[file.Repository]; !ok {
				repositoryFiles[file.Repository] = make(map[string]*AggregatedFile)
			}

			// We got files for this repo. Let's check if it is the first time we're
			// seeing this file. In that case we can just go ahead and add it.
			if _, ok := repositoryFiles[file.Repository][file.Path]; !ok {
				repositoryFiles[file.Repository][file.Path] = &AggregatedFile{
					Name:       file.Name,
					Path:       file.Path,
					Filetype:   file.Filetype,
					DurationMs: file.DurationMs,
				}
				continue
			}

			// We've seen this file and repository before. We'll have to merge them.
			prevFile := repositoryFiles[file.Repository][file.Path]
			prevFile.DurationMs += file.DurationMs
		}
	}
	return repositoryFiles
}

// You could work on several different repositories during one
// coding session. This function groups the work by repository
func repositoriesFromSessions(sessions Sessions) []Repository {
	repositoryFiles := repositoryFileMap(sessions)
	repositories := make([]Repository, 0)

	for repositoryName, filePointers := range repositoryFiles {
		var durationMs int64 = 0
		files := make([]AggregatedFile, 0)

		for _, filePointer := range filePointers {
			files = append(files, *filePointer)
			durationMs += filePointer.DurationMs
		}

		repository := Repository{
			Name:       repositoryName,
			Files:      files,
			DurationMs: durationMs,
		}

		repositories = append(repositories, repository)
	}

	return repositories
}
