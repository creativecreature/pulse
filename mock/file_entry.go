package mock

import "io/fs"

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
