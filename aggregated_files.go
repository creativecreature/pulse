package domain

// AggregatedFiles represents a slice of files that has been
// aggregated for a given time period. Raw sessions are aggregated
// by day. Daily sessions are aggregated by week, month, and year.
type AggregatedFiles []AggregatedFile

// createPathFileMap takes a slice of aggregated files and produces a map, where
// the file path is used as the key and the file itself serves as the value.
func createPathFileMap(files AggregatedFiles) map[string]AggregatedFile {
	pathFileMap := make(map[string]AggregatedFile)
	for _, f := range files {
		pathFileMap[f.Path] = f
	}
	return pathFileMap
}

// merge takes two slices of aggregated files, merges them, and returns the result.
func (a AggregatedFiles) merge(b AggregatedFiles) AggregatedFiles {
	mergedFiles := make([]AggregatedFile, 0)
	aFileMap := createPathFileMap(a)
	bFileMap := createPathFileMap(b)

	// Add files that are unique for a and merge collisions.
	for _, aFile := range a {
		if bFile, ok := bFileMap[aFile.Path]; !ok {
			mergedFiles = append(mergedFiles, aFile)
		} else {
			mergedFiles = append(mergedFiles, aFile.merge(bFile))
		}
	}

	// Add the files that are unique for b.
	for _, bFile := range b {
		if _, ok := aFileMap[bFile.Path]; !ok {
			mergedFiles = append(mergedFiles, bFile)
		}
	}

	return mergedFiles
}
