package domain

type Repositories []Repository

// You could work on several different repositories during one
// coding session. This function groups the work by repository
func repositoriesFromSessions(sessions Sessions) []Repository {
	repositoryFiles := repositoryFileMap(sessions)
	repositories := make([]Repository, 0)

	for repositoryName, filePointers := range repositoryFiles {
		var durationMs int64 = 0
		files := make([]AggregatedFile, 0)

		for _, filePointer := range filePointers {
			files = append(files, filePointer)
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

// Merges the repositories of two aggregated sessions
func (a Repositories) merge(b Repositories) Repositories {
	mergedRepositories := make([]Repository, 0)
	aRepoMap := make(map[string]Repository)
	bRepoMap := make(map[string]Repository)
	for _, aRepo := range a {
		aRepoMap[aRepo.Name] = aRepo
	}
	for _, bRepo := range b {
		bRepoMap[bRepo.Name] = bRepo
	}

	// Add repos that are unique for a and merge collisions
	for _, aRepo := range a {
		bRepo, ok := bRepoMap[aRepo.Name]
		if !ok {
			mergedRepositories = append(mergedRepositories, aRepo)
			continue
		}
		mergedRepositories = append(mergedRepositories, aRepo.merge(bRepo))
	}

	// The merging is done at this point. Here we'll add the repositories that
	// are unique to the new session
	for _, newRepo := range b {
		if _, ok := aRepoMap[newRepo.Name]; !ok {
			mergedRepositories = append(mergedRepositories, newRepo)
		}
	}

	return mergedRepositories
}
