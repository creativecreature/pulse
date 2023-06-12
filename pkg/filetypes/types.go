package filetypes

import (
	"errors"
	"strings"
)

var ErrUnrecognizedFileExtension = errors.New("could not extract a filetype from the filename")

// Get tries to extract the filetype from a filename.
func Type(filename string) (string, error) {
	// Start by checking if it is a special file
	file, ok := filenameTypeMap[strings.Split(filename, ".")[0]]
	if ok {
		return file, nil
	}

	// This will return an empty string (zero value) if we don't have a match.
	parts := strings.Split(filename, ".")
	fileExtension := parts[len(parts)-1]
	filetype, ok := extensionTypeMap[fileExtension]

	if !ok {
		return "", ErrUnrecognizedFileExtension
	}

	return filetype, nil
}
