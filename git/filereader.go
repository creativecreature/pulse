package git

import (
	"io/fs"
	"os"
	"path/filepath"
)

// filereader implements the Reader interface and adds functionality
// for reading files from the underlying filesystem.
type filereader struct{}

// Dir is a wrapper around filepath.Dir.
func (f filereader) Dir(path string) string {
	return filepath.Dir(path)
}

// ReadDir is a wrapper around os.ReadDir.
func (f filereader) ReadDir(dir string) ([]fs.DirEntry, error) {
	return os.ReadDir(dir)
}

// ReadFile is a wrapper around os.ReadFile.
func (f filereader) ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

// IsFile is a wrapper around os.Stat. It returns the negation (!)
// of calling IsDir() on the FileInfo that was returned by os.Stat.
func (f filereader) IsFile(path string) bool {
	fileInfo, err := os.Stat(path)
	return err == nil && !fileInfo.IsDir()
}
