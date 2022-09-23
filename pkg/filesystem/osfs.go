package filesystem

import (
	"io/fs"
	"os"
	"path/filepath"
)

// OsFS implements the FileSystem with functions that will be using the underlying OS
type OsFS struct {
	FileSystem
}

func (f OsFS) Dir(path string) string {
	return filepath.Dir(path)
}

func (f OsFS) ReadDir(dir string) ([]fs.DirEntry, error) {
	return os.ReadDir(dir)
}

func (f OsFS) ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

func (f OsFS) IsFile(path string) bool {
	fileInfo, err := os.Stat(path)
	return err == nil && !fileInfo.IsDir()
}
