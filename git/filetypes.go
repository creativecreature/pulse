package git

import (
	"errors"
	"strings"
)

var ErrUnrecognizedFileExtension = errors.New("could not extract a filetype from the filename")

// Map of file extensions to filetypes.
var extensions = map[string]string{
	"C":      "c++",
	"H":      "c++",
	"c":      "c",
	"clj":    "clojure",
	"cljs":   "clojure",
	"cpp":    "c++",
	"css":    "css",
	"csv":    "csv",
	"edn":    "clojure",
	"ex":     "elixir",
	"exs":    "elixir",
	"go":     "go",
	"h":      "c",
	"hpp":    "c++",
	"hs":     "haskell",
	"html":   "html",
	"java":   "java",
	"js":     "javascript",
	"json":   "json",
	"jsx":    "javascript",
	"kt":     "kotlin",
	"less":   "css",
	"lua":    "lua",
	"md":     "markdown",
	"mdx":    "markdown",
	"ml":     "ocaml",
	"mli":    "ocaml",
	"php":    "php",
	"py":     "python",
	"rb":     "ruby",
	"rs":     "rust",
	"scss":   "css",
	"sh":     "bash",
	"swift":  "swift",
	"templ":  "go",
	"tf":     "terraform",
	"tfvars": "terraform",
	"toml":   "toml",
	"ts":     "typescript",
	"tsx":    "typescript",
	"vim":    "vimscript",
	"yaml":   "yaml",
	"yml":    "yaml",
}

// Special files that we track time for.
var specialFiles = map[string]string{
	"Makefile":       "Makefile",
	"Dockerfile":     "Docker",
	"docker-compose": "Docker",
}

// Filetype extracts the filetype from a filename.
func Filetype(filename string) (string, error) {
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
