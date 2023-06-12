package filesystem

import "io/fs"

type GitFile interface {
	Name() string
	Filetype() string
	Repository() string
	Path() string
}

type Filesystem interface {
	Dir(string) string
	ReadDir(string) ([]fs.DirEntry, error)
	ReadFile(string) ([]byte, error)
	IsFile(string) bool
}
