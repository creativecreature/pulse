package domain

type AggregatedFiles []AggregatedFile

func createPathFileMap(files AggregatedFiles) map[string]AggregatedFile {
	pathFileMap := make(map[string]AggregatedFile)
	for _, f := range files {
		pathFileMap[f.Path] = f
	}
	return pathFileMap
}

// merge merges two AggregatedFile slices
func (a AggregatedFiles) merge(b AggregatedFiles) AggregatedFiles {
	mergedFiles := make([]AggregatedFile, 0)
	aFileMap := createPathFileMap(a)
	bFileMap := createPathFileMap(b)

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
