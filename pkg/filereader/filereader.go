package filereader

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

type OSFileReader struct{}

func (f OSFileReader) Dir(path string) string {
	return filepath.Dir(path)
}

func (f OSFileReader) ReadDir(dir string) ([]fs.DirEntry, error) {
	return os.ReadDir(dir)
}

func (f OSFileReader) ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

func (f OSFileReader) IsFile(path string) bool {
	fileInfo, err := os.Stat(path)
	return err == nil && !fileInfo.IsDir()
}
