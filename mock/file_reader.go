package mock

import (
	"code-harvest.conner.dev/domain"
)

// FileReader is a mock for the FileReader interface
type FileReader struct {
	file domain.GitFile
}

func (f *FileReader) GitFile(path string) (domain.GitFile, error) {
	return f.file, nil
}

func (f *FileReader) SetFile(file domain.GitFile) {
	f.file = file
}

func NewFileReader() *FileReader {
	return &FileReader{}
}
