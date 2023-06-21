package filereader

import (
	"io/fs"
	"os"
	"path/filepath"
)

// Interface for the underlying reader
type Reader interface {
	Dir(string) string
	ReadDir(string) ([]fs.DirEntry, error)
	ReadFile(string) ([]byte, error)
	IsFile(string) bool
}

type reader struct{}

func (f reader) Dir(path string) string {
	return filepath.Dir(path)
}

func (f reader) ReadDir(dir string) ([]fs.DirEntry, error) {
	return os.ReadDir(dir)
}

func (f reader) ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

func (f reader) IsFile(path string) bool {
	fileInfo, err := os.Stat(path)
	return err == nil && !fileInfo.IsDir()
}
