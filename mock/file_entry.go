package mock

import "io/fs"

// FileEntry is a mock of the fs.FileEntry interface
type FileEntry struct {
	fs.DirEntry
	Filename    string
	IsDirectory bool
}

func (f FileEntry) Name() string {
	return f.Filename
}

func (f FileEntry) IsDir() bool {
	return f.IsDirectory
}

func NewFileEntry(filename string, isDirectory bool) FileEntry {
	return FileEntry{
		Filename:    filename,
		IsDirectory: isDirectory,
	}
}
