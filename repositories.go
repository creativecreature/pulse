package pulse

// Repositories represents a list of git repositories.
type Repositories []Repository

// repositories take a slice of sessions and returns the repositories.
func repositories(sessions Sessions) Repositories {
	namesWithFiles := repositoryNamesWithFiles(sessions)
	repos := make(Repositories, 0, len(namesWithFiles))

	for repositoryName, filepathFile := range namesWithFiles {
		files := make(AggregatedFiles, 0, len(filepathFile))
		repo := Repository{Name: repositoryName, Files: files}
		for _, file := range filepathFile {
			repo.Files = append(repo.Files, file)
			repo.DurationMs += file.DurationMs
		}
		repos = append(repos, repo)
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

// merge takes two lists of repositories, merges them, and returns the result.
func (r Repositories) merge(b Repositories) Repositories {
	aNames, bNames := repositoriesByName(r), repositoriesByName(b)
	allNames := make(map[string]bool)
	for name := range aNames {
		allNames[name] = true
	}
	for name := range bNames {
		allNames[name] = true
	}

	mergedRepositories := make([]Repository, 0)
	for name := range allNames {
		aRepo := aNames[name]
		bRepo := bNames[name]
		mergedRepositories = append(mergedRepositories, aRepo.merge(bRepo))
	}
	return mergedRepositories
}
