package domain

type Repositories []Repository

func (previousRepositories Repositories) Merge(newRepositories Repositories) Repositories {
	prevReposMap := make(map[string]Repository)
	newReposMap := make(map[string]Repository)
	for _, repository := range previousRepositories {
		prevReposMap[repository.Name] = repository
	}
	for _, repository := range newRepositories {
		newReposMap[repository.Name] = repository
	}

	mergedRepositories := make([]Repository, 0)
	for _, prevRepo := range previousRepositories {
		// This repository haven't been worked on in the new session. We'll just
		// add it to the final slice
		newRepo, ok := newReposMap[prevRepo.Name]
		if !ok {
			mergedRepositories = append(mergedRepositories, prevRepo)
			continue
		}

		// This repository has been worked on in both sessions. We'll have to merge them
		mergedFiles := prevRepo.Files.Merge(newRepo.Files)
		mergedRepository := Repository{
			Name:       prevRepo.Name,
			DurationMs: prevRepo.DurationMs + newRepo.DurationMs,
			Files:      mergedFiles,
		}
		mergedRepositories = append(mergedRepositories, mergedRepository)
	}

	for _, newRepo := range newRepositories {
		// We have already handled the merging in the loop above. Here we'll just
		// add the new repositories which haven't been worked on in the previous
		// session.
		if _, ok := prevReposMap[newRepo.Name]; !ok {
			mergedRepositories = append(mergedRepositories, newRepo)
		}
	}

	return mergedRepositories
}
