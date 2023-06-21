package domain

type AggregatedFiles []AggregatedFile

// merge merges two AggregatedFile slices
func (a AggregatedFiles) merge(b AggregatedFiles) AggregatedFiles {
	mergedFiles := make([]AggregatedFile, 0)
	aFileMap := make(map[string]AggregatedFile)
	bFileMap := make(map[string]AggregatedFile)
	for _, f := range a {
		aFileMap[f.Path] = f
	}
	for _, f := range b {
		bFileMap[f.Path] = f
	}

	// Add files that are unique for a and merge collisions
	for _, aFile := range a {
		if bFile, ok := bFileMap[aFile.Path]; !ok {
			mergedFiles = append(mergedFiles, aFile)
		} else {
			mergedFiles = append(mergedFiles, aFile.merge(bFile))
		}
	}

	// Add the files that are unique for b
	for _, bFile := range b {
		if _, ok := aFileMap[bFile.Path]; !ok {
			mergedFiles = append(mergedFiles, bFile)
		}
	}

	return mergedFiles
}
