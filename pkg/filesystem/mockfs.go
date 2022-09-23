package filesystem

import (
	"errors"
	"io/fs"
)

// MockFS implements the FileSystem with functions that makes it easy to
// test code that accesses the underlying os and file system.
type MockFS struct {
	DirectoryIndex int
	Directories    []string
	Entries        map[string][]fs.DirEntry
	FileContents   map[string][]byte
	FileSystem
}

func (f *MockFS) Dir(_ string) string {
	if f.DirectoryIndex > len(f.Directories)-1 {
		return ""
	}
	dir := f.Directories[f.DirectoryIndex]
	f.DirectoryIndex++
	return dir
}

func (f *MockFS) ReadDir(dir string) ([]fs.DirEntry, error) {
	entries, ok := f.Entries[dir]
	if !ok {
		return nil, errors.New("no entries for dir")
	}
	return entries, nil
}

func (f *MockFS) ReadFile(filename string) ([]byte, error) {
	fileContent, ok := f.FileContents[filename]
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
