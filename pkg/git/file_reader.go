package git

import (
	"io/fs"
	"os"
	"path/filepath"
)

type FileReader interface {
	Dir(string) string
	ReadDir(string) ([]fs.DirEntry, error)
	ReadFile(string) ([]byte, error)
	IsFile(string) bool
}

type fileReader struct{}

func (f fileReader) Dir(path string) string {
	return filepath.Dir(path)
}

func (f fileReader) ReadDir(dir string) ([]fs.DirEntry, error) {
	return os.ReadDir(dir)
}

func (f fileReader) ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

func (f fileReader) IsFile(path string) bool {
	fileInfo, err := os.Stat(path)
	return err == nil && !fileInfo.IsDir()
}
