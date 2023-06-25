package domain

type Repositories []Repository

// repositories take a slice of sessions and returns the repositories.
func repositories(sessions Sessions) []Repository {
	filesByRepo := repositoryPathFile(sessions)
	repos := make([]Repository, 0)

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
		repos = append(repos, repository)
	}

	return repos
}

// repositoriesByName takes a slice of repositories and returns a map
// where the repository name is the key and the repository the value.
func repositoriesByName(repos Repositories) map[string]Repository {
	nameRepoMap := make(map[string]Repository)
	for _, repo := range repos {
		nameRepoMap[repo.Name] = repo
	}
	return nameRepoMap
}

// Merge takes two slices of repositories, merges them, and returns the result.
func (a Repositories) merge(b Repositories) Repositories {
	mergedRepositories := make([]Repository, 0)
	aRepoMap, bRepoMap := repositoriesByName(a), repositoriesByName(b)

	// Add repos that are unique for a and merge collisions.
	for _, aRepo := range a {
		if bRepo, ok := bRepoMap[aRepo.Name]; !ok {
			mergedRepositories = append(mergedRepositories, aRepo)
		} else {
			mergedRepositories = append(mergedRepositories, aRepo.merge(bRepo))
		}
	}

	// The merging is done at this point. Here we'll add the
	// repositories that are unique to the new session.
	for _, newRepo := range b {
		if _, ok := aRepoMap[newRepo.Name]; !ok {
			mergedRepositories = append(mergedRepositories, newRepo)
		}
	}

	return mergedRepositories
}
