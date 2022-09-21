package filesystem

import (
	"errors"
	"io/fs"
)

// Implements the FileSystem with functions that makes it easy to
// test code that accesses the underlying os and file system.
type MockFS struct {
	DirectoryIndex int
	Directories    []string
	Entries        map[string][]fs.DirEntry
	FileContents   map[string][]byte
	FileSystem
}

func (f *MockFS) Dir(path string) string {
	if f.DirectoryIndex > len(f.Directories)-1 {
		return ""
	}
	dir := f.Directories[f.DirectoryIndex]
	f.DirectoryIndex++
	return dir
}

func (o *MockFS) ReadDir(dir string) ([]fs.DirEntry, error) {
	entries, ok := o.Entries[dir]
	if !ok {
		return nil, errors.New("no entries for dir")
	}
	return entries, nil
}

func (o *MockFS) ReadFile(filename string) ([]byte, error) {
	fileContent, ok := o.FileContents[filename]
	if !ok {
		return nil, errors.New("no content for this filename")
	}
	return fileContent, nil
}

type MockFileEntry struct {
	fs.DirEntry
	Filename    string
	IsDirectory bool
}

func (f MockFileEntry) Name() string {
	return f.Filename
}

func (f MockFileEntry) IsDir() bool {
	return f.IsDirectory
}
