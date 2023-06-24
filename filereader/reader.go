package filereader

import (
	"io/fs"
	"os"
	"path/filepath"
)

// Reader is an interface for the underlying reader. It is an abstraction that
// allows for easier testing
type Reader interface {
	Dir(string) string
	ReadDir(string) ([]fs.DirEntry, error)
	ReadFile(string) ([]byte, error)
	IsFile(string) bool
}

// reader implements the Reader interface and adds functionality for reading
// files from the underlying filesystem
type reader struct{}

// Dir is a wrapper around filepath.Dir
func (f reader) Dir(path string) string {
	return filepath.Dir(path)
}

// ReadDir is a wrapper around os.ReadDir
func (f reader) ReadDir(dir string) ([]fs.DirEntry, error) {
	return os.ReadDir(dir)
}

// ReadFile is a wrapper around os.ReadFile
func (f reader) ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

// IsFile is a wrapper around os.Stat. It returns the negation (!) of calling
// IsDir() on the FileInfo that was returned by os.Stat
func (f reader) IsFile(path string) bool {
	fileInfo, err := os.Stat(path)
	return err == nil && !fileInfo.IsDir()
}
