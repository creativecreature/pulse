package mock

import (
	"fmt"
	"io/fs"
	"strings"
)

type Reader struct {
	DirectoryIndex int
	Directories    []string
	Entries        map[string][]fs.DirEntry
	FileContents   map[string][]byte
}

func (f *Reader) Dir(_ string) string {
	if f.DirectoryIndex > len(f.Directories)-1 {
		return ""
	}
	dir := f.Directories[f.DirectoryIndex]
	f.DirectoryIndex++
	return dir
}

func (f *Reader) ReadDir(dir string) ([]fs.DirEntry, error) {
	entries, ok := f.Entries[dir]
	if !ok {
		return nil, fmt.Errorf("no entries for dir: %s", dir)
	}
	return entries, nil
}

func (f *Reader) ReadFile(filename string) ([]byte, error) {
	fileContent, ok := f.FileContents[filename]
	if !ok {
		return nil, fmt.Errorf("no file content for file: %s", filename)
	}
	return fileContent, nil
}

func (f *Reader) IsFile(filename string) bool {
	return true
}

func (f *Reader) Filename(path string) string {
	s := strings.Split(path, "/")
	return s[len(s)-1]
}
