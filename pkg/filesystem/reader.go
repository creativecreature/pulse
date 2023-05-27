package filesystem

import (
	"errors"
	"io/fs"
	"path/filepath"
)

var (
	ErrEmptyPath                       = errors.New("path is empty string")
	ErrPathNotAFile                    = errors.New("the path is not a file")
	ErrFileNotUnderSourceControl       = errors.New("the files does not reside within a repository")
	ErrReachedRoot                     = errors.New("we reached the root without finding a .git file or folder")
	ErrParseRepoPath                   = errors.New("failed to parse repository path")
	ErrParseBareRepoPath               = errors.New("failed to parse bare repository path")
	ErrRepositoryDirectoryNameMismatch = errors.New("could not extract relative path in repo")
)

type Filesystem interface {
	Dir(string) string
	ReadDir(string) ([]fs.DirEntry, error)
	ReadFile(string) ([]byte, error)
	IsFile(string) bool
}

type FileReader struct {
	fsys Filesystem
}

func NewReader(fsys Filesystem) FileReader {
	return FileReader{fsys}
}

func (f FileReader) Read(path string) (File, error) {
	if path == "" {
		return file{}, ErrEmptyPath
	}

	// It could be a temporary buffer or directory.
	if !f.fsys.IsFile(path) {
		return file{}, ErrPathNotAFile
	}

	// When I aggregate the data I do it on a per project basis. Therefore, if this
	// is just a one-off edit of some configuration file I won't track time for it.
	repositoryName, err := f.RepositoryName(path)
	if err != nil {
		return file{}, err
	}

	filename := filepath.Base(path)

	// Tries to get the filetype from either the file extension or name.
	ft, err := Filetype(filename)
	if err != nil {
		return file{}, err
	}

	fileMetaData := file{filename, ft, repositoryName}
	return fileMetaData, nil
}
