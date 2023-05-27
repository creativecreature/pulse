package mock

import (
	"errors"

	"code-harvest.conner.dev/internal/filereader"
)

type fileReader struct {
	file filereader.File
}

func (f *fileReader) Read(path string) (filereader.File, error) {
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
