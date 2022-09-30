package file

import (
	"errors"
	"path/filepath"
	"time"

	"code-harvest.conner.dev/pkg/filesystem"
	"code-harvest.conner.dev/pkg/filetype"
	"code-harvest.conner.dev/pkg/git"
)

var fs = filesystem.OsFS{}

var ErrNotAFile = errors.New("path is dir or temporary buffer")

type File struct {
	Name       string `bson:"name"`
	Repository string `bson:"repository"`
	Path       string `bson:"path"`
	Filetype   string `bson:"filetype"`
	DurationMs int64  `bson:"duration_ms"`
	OpenedAt   int64  `bson:"-"`
	ClosedAt   int64  `bson:"-"`
}

func New(path string) (*File, error) {
	openedAt := time.Now().UTC().UnixMilli()

	// It could be a temporary buffer
	if !fs.IsFile(path) {
		return nil, ErrNotAFile
	}

	// If the file isn't in a repository I don't want to track time for it.
	repository, err := git.GetRepositoryFromPath(path)
	if err != nil {
		return nil, err
	}

	// Tries to get the relative path to the file from the repository root.
	relativePathInRepo, err := git.GetRelativePathFromRepo(path, repository)
	if err != nil {
		return nil, err
	}

	// Extract the name of the file.
	name := filepath.Base(relativePathInRepo)

	// Tries to get the filetype from either the file extension or name.
	ft, err := filetype.Get(name)
	if err != nil {
		return nil, err
	}

	file := File{
		Name:       name,
		Repository: repository,
		Path:       path,
		Filetype:   ft,
		OpenedAt:   openedAt,
		ClosedAt:   0,
	}

	return &file, nil
}
