package pulse

// AggregatedFiles represents a slice of files that has been
// aggregated for a given time period (day, week, month, year).
type AggregatedFiles []AggregatedFile

// createPathFileMap takes a slice of aggregated files and produces a map,
// where the path is used as the key and the file itself serves as the value.
func createPathFileMap(files AggregatedFiles) map[string]AggregatedFile {
	pathFileMap := make(map[string]AggregatedFile)
	for _, f := range files {
		pathFileMap[f.Path] = f
	}
	return pathFileMap
}

func (a AggregatedFiles) merge(b AggregatedFiles) AggregatedFiles {
	aFileMap := createPathFileMap(a)
	bFileMap := createPathFileMap(b)

	allPaths := make(map[string]struct{})
	for path := range aFileMap {
		allPaths[path] = struct{}{}
	}
	for path := range bFileMap {
		allPaths[path] = struct{}{}
	}

	mergedFiles := make([]AggregatedFile, 0, len(allPaths))
	for path := range allPaths {
		aFile := aFileMap[path]
		bFile := bFileMap[path]
		mergedFiles = append(mergedFiles, aFile.merge(bFile))
	}
	return mergedFiles
}
