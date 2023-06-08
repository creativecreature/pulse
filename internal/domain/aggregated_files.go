package domain

type AggregatedFiles []AggregatedFile

func (prevFiles AggregatedFiles) Merge(newFiles AggregatedFiles) AggregatedFiles {
	prevFilesMap := make(map[string]AggregatedFile)
	newFilesMap := make(map[string]AggregatedFile)
	for _, file := range prevFiles {
		prevFilesMap[file.Path] = file
	}
	for _, file := range newFiles {
		newFilesMap[file.Path] = file
	}

	mergedFiles := make([]AggregatedFile, 0)
	for _, prevFile := range prevFiles {
		// This file haven't been worked on in the new session. We'll just
		// add it to the final slice
		newFile, ok := newFilesMap[prevFile.Path]
		if !ok {
			mergedFiles = append(mergedFiles, prevFile)
			continue
		}

		mergedFile := AggregatedFile{
			Name:       prevFile.Name,
			Path:       prevFile.Path,
			Filetype:   prevFile.Filetype,
			DurationMs: prevFile.DurationMs + newFile.DurationMs,
		}
		mergedFiles = append(mergedFiles, mergedFile)
	}

	for _, newFile := range newFiles {
		// We have already handled the merging in the loop above. Here we'll just
		// add the new file which haven't been worked on in the previous session.
		if _, ok := prevFilesMap[newFile.Path]; !ok {
			mergedFiles = append(mergedFiles, newFile)
		}
	}

	return mergedFiles
}
