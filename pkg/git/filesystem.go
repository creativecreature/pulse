package git

import "io/fs"

type Filesystem interface {
	Dir(string) string
	ReadDir(string) ([]fs.DirEntry, error)
	ReadFile(string) ([]byte, error)
}
