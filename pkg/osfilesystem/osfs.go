package osfilesystem

import (
	"io/fs"
	"os"
	"path/filepath"
)

type osfs struct{}

func New() osfs {
	return osfs{}
}

func (f osfs) Dir(path string) string {
	return filepath.Dir(path)
}

func (f osfs) ReadDir(dir string) ([]fs.DirEntry, error) {
	return os.ReadDir(dir)
}

func (f osfs) ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}
