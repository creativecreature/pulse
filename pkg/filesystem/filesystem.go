package filesystem

import (
	"io/fs"
	"os"
	"path/filepath"
)

type filesystem struct{}

func New() filesystem {
	return filesystem{}
}

func (f filesystem) Dir(path string) string {
	return filepath.Dir(path)
}

func (f filesystem) ReadDir(dir string) ([]fs.DirEntry, error) {
	return os.ReadDir(dir)
}

func (f filesystem) ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

func (f filesystem) IsFile(path string) bool {
	fileInfo, err := os.Stat(path)
	return err == nil && !fileInfo.IsDir()
}
