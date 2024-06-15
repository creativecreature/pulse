package pulse

import "cmp"

// GitFile represents a file within a git repository.
type GitFile struct {
	Name       string
	Filetype   string
	Repository string
	Path       string
}

// File represents a file that has been aggregated
// for a given time period (day, week, month, year).
type File struct {
	Name       string `bson:"name"`
	Path       string `bson:"path"`
	Filetype   string `bson:"filetype"`
	DurationMs int64  `bson:"duration_ms"`
}

// merge takes two files, merges them, and returns the result.
func (a File) merge(b File) File {
	return File{
		Name:       cmp.Or(a.Name, b.Name),
		Path:       cmp.Or(a.Path, b.Path),
		Filetype:   cmp.Or(a.Filetype, b.Filetype),
		DurationMs: a.DurationMs + b.DurationMs,
	}
}

// Files represents a slice of files that has been aggregated
// for a given time period (day, week, month, year).
type Files []File

// createPathFileMap takes a slice of files and produces a map, where the
// filepath is used as the key and the file itself serves as the value.
func createPathFileMap(files Files) map[string]File {
	pathFileMap := make(map[string]File)
	for _, f := range files {
		pathFileMap[f.Path] = f
	}
	return pathFileMap
}

// merge takes two slices of files, merges them, and returns the result.
func (a Files) merge(b Files) Files {
	aFileMap := createPathFileMap(a)
	bFileMap := createPathFileMap(b)

	allPaths := make(map[string]struct{})
	for path := range aFileMap {
		allPaths[path] = struct{}{}
	}
	for path := range bFileMap {
		allPaths[path] = struct{}{}
	}

	mergedFiles := make([]File, 0, len(allPaths))
	for path := range allPaths {
		aFile := aFileMap[path]
		bFile := bFileMap[path]
		mergedFiles = append(mergedFiles, aFile.merge(bFile))
	}
	return mergedFiles
}
