package app

import (
	"errors"
	"os"
	"path/filepath"

	"code-harvest.conner.dev/pkg/filetype"
	"code-harvest.conner.dev/pkg/git"
)

var (
	ErrEmptyPath                 = errors.New("path is empty string")
	ErrPathNotAFile              = errors.New("the path is not a file")
	ErrFileNotUnderSourceControl = errors.New("the files does not reside within a repository")
)

type FileMetadata struct {
	Filename       string
	Filetype       string
	RepositoryName string
}

type MetadataReader interface {
	Read(uri string) (FileMetadata, error)
}

// Extracts metadata from a file
type FileReader struct{}

func isFile(path string) bool {
	fileInfo, err := os.Stat(path)
	return err == nil && !fileInfo.IsDir()
}

func (f FileReader) Read(path string) (FileMetadata, error) {
	if path == "" {
		return FileMetadata{}, ErrEmptyPath
	}

	// It could be a temporary buffer or directory.
	if !isFile(path) {
		return FileMetadata{}, ErrPathNotAFile
	}

	// When I aggregate the data I do it on a per project basis. Therefore, if this
	// is just a one-off edit of some configuration file I won't track time for it.
	repositoryName, err := git.GetRepositoryNameFromPath(path)
	if err != nil {
		return FileMetadata{}, err
	}

	filename := filepath.Base(path)

	// Tries to get the filetype from either the file extension or name.
	ft, err := filetype.Get(filename)
	if err != nil {
		return FileMetadata{}, err
	}

	fileMetaData := FileMetadata{
		Filename:       filename,
		Filetype:       ft,
		RepositoryName: repositoryName,
	}

	return fileMetaData, nil
}
