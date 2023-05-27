package mock

import (
	"errors"

	"code-harvest.conner.dev/pkg/filesystem"
)

type fileReader struct {
	file filesystem.File
}

func (f *fileReader) Read(path string) (filesystem.File, error) {
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
