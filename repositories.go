package pulse

// Repositories represents a list of git repositories.
type Repositories []Repository

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
