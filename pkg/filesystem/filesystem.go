package filesystem

import (
	"io/fs"
)

type FileSystem interface {
	Dir(string) string
	ReadDir(string) ([]fs.DirEntry, error)
	ReadFile(string) ([]byte, error)
	IsFile(path string) bool
}
