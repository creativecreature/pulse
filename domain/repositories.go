package domain

type Repositories []Repository

// You could work on several different repositories during one
// coding session. This function groups the work by repository
func sessionRepositories(sessions Sessions) []Repository {
	filesByRepo := aggregateFilesByRepo(sessions)
	repositories := make([]Repository, 0)

	for repositoryName, filenameFileMap := range filesByRepo {
		var durationMs int64 = 0
		files := make([]AggregatedFile, 0)
		for _, file := range filenameFileMap {
			files = append(files, file)
			durationMs += file.DurationMs
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

func createNameRepoMap(repos Repositories) map[string]Repository {
	nameRepoMap := make(map[string]Repository)
	for _, repo := range repos {
		nameRepoMap[repo.Name] = repo
	}
	return nameRepoMap
}

// Merges the repositories of two aggregated sessions
func (a Repositories) merge(b Repositories) Repositories {
	mergedRepositories := make([]Repository, 0)
	aRepoMap, bRepoMap := createNameRepoMap(a), createNameRepoMap(b)

	// Add repos that are unique for a and merge collisions
	for _, aRepo := range a {
		if bRepo, ok := bRepoMap[aRepo.Name]; !ok {
			mergedRepositories = append(mergedRepositories, aRepo)
		} else {
			mergedRepositories = append(mergedRepositories, aRepo.merge(bRepo))
		}
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
