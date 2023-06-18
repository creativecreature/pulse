package mock

import (
	"errors"

	"code-harvest.conner.dev/pkg/git"
)

// fileReader is a mock for the FileReader interface
type fileReader struct {
	file git.File
}

func (f *fileReader) GitFile(path string) (git.File, error) {
	if f.file == nil {
		return File{}, errors.New("metadata is nil")
	}
	return f.file, nil
}

func (f *fileReader) SetFile(file File) {
	f.file = file
}

func (f *fileReader) ClearFile() {
	f.file = nil
}

func NewReader() *fileReader {
	return &fileReader{}
}
