package domain

// Repository represents all work that has been done in a repository during a day
type Repository struct {
	Name       string      `bson:"name"`
	Files      []DailyFile `bson:"files"`
	DurationMs int64       `bson:"duration_ms"`
}

// repositores takes all the temporary sessions for a given day and returns all
// the repositories, and files within those repositories, that we've worked on.
func repositories(sessions []Session) []Repository {
	// The coding sessions may have opened files in different repositories.
	// Therefore, we'll have to aggregate all the files together by repository
	// and filename.
	repositoryFileMap := make(map[string]map[string]*DailyFile)
	for _, tempSession := range sessions {
		for _, file := range tempSession.Files {
			// Check if it is the first time we're seeing this repo. In that case
			// we'll initialize the map.
			if _, ok := repositoryFileMap[file.Repository]; !ok {
				repositoryFileMap[file.Repository] = make(map[string]*DailyFile)
			}

			// We got files for this repo. Let's check if it is the first time we're
			// seeing this file. In that case we can just go ahead and add it.
			if _, ok := repositoryFileMap[file.Repository][file.Path]; !ok {
				repositoryFileMap[file.Repository][file.Path] = &DailyFile{
					Name:       file.Name,
					Path:       file.Path,
					Filetype:   file.Filetype,
					DurationMs: file.DurationMs,
				}
				continue
			}

			// We've seen this file and repository before. We'll have to merge them.
			prevFile := repositoryFileMap[file.Repository][file.Path]
			prevFile.DurationMs += file.DurationMs
		}
	}

	// At this point we've merged all the files that could have been edited
	// during different sessions and saved them in our repository/filename map.
	// We can now turn the maps into a slice of repositories.
	repositories := make([]Repository, 0)
	for repositoryName, filePointers := range repositoryFileMap {
		var durationMs int64 = 0
		files := make([]DailyFile, 0)
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
