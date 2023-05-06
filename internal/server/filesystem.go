package server

import (
	"io/fs"
	"os"
	"path/filepath"
)

type filesystem struct{}

func (f filesystem) Dir(path string) string {
	return filepath.Dir(path)
}

func (f filesystem) ReadDir(dir string) ([]fs.DirEntry, error) {
	return os.ReadDir(dir)
}

func (f filesystem) ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}
