package filetypes

import (
	"errors"
	"strings"
)

var ErrUnrecognizedFileExtension = errors.New("could not extract a filetype from the filename")

// Map of file extensions to filetypes.
var extensions = map[string]string{
	"go":    "go",
	"js":    "javascript",
	"jsx":   "javascript",
	"ts":    "typescript",
	"tsx":   "typescript",
	"lua":   "lua",
	"md":    "markdown",
	"mdx":   "markdown",
	"css":   "css",
	"less":  "css",
	"scss":  "css",
	"sh":    "bash",
	"vim":   "vimscript",
	"yml":   "yaml",
	"yaml":  "yaml",
	"json":  "json",
	"toml":  "toml",
	"hpp":   "c++",
	"H":     "c++",
	"C":     "c++",
	"cpp":   "c++",
	"c":     "c",
	"h":     "c",
	"rs":    "rust",
	"py":    "python",
	"java":  "java",
	"kt":    "kotlin",
	"swift": "swift",
	"php":   "php",
	"rb":    "ruby",
	"html":  "html",
	"hs":    "haskell",
}

// Special files that we track time for.
var specialFiles = map[string]string{
	"Makefile":       "Makefile",
	"Dockerfile":     "Docker",
	"docker-compose": "Docker",
}

// Type extracts the filetype from a filename.
func Type(filename string) (string, error) {
	// Start by checking if it is a special file.
	if file, ok := specialFiles[strings.Split(filename, ".")[0]]; ok {
		return file, nil
	}

	// This will return an empty string (zero value)
	// if we don't have a match.
	parts := strings.Split(filename, ".")
	fileExtension := parts[len(parts)-1]
	if filetype, ok := extensions[fileExtension]; ok {
		return filetype, nil
	}

	return "", ErrUnrecognizedFileExtension
}
