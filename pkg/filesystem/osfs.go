package filesystem

import (
	"io/fs"
	"os"
	"path/filepath"
)

// Implements the FileSystem with functions that will be using the underlying OS
type OsFS struct {
	FileSystem
}

func (f OsFS) Dir(path string) string {
	return filepath.Dir(path)
}

func (o OsFS) ReadDir(dir string) ([]fs.DirEntry, error) {
	return os.ReadDir(dir)
}

func (o OsFS) ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

func (o OsFS) IsFile(path string) bool {
	fileInfo, err := os.Stat(path)
	return err == nil && !fileInfo.IsDir()
}
