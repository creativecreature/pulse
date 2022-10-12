package git

import (
	"io/fs"
	"os"
	"path/filepath"
)

type osFS struct {
	FileSystem
}

func (f osFS) Dir(path string) string {
	return filepath.Dir(path)
}

func (f osFS) ReadDir(dir string) ([]fs.DirEntry, error) {
	return os.ReadDir(dir)
}

func (f osFS) ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

func (f osFS) IsFile(path string) bool {
	fileInfo, err := os.Stat(path)
	return err == nil && !fileInfo.IsDir()
}
