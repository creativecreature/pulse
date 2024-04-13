package mock

import (
	"github.com/creativecreature/pulse"
)

// FileReader is a mock implementation of the domain.FileReader interface.
type FileReader struct {
	file pulse.GitFile
}

func (f *FileReader) GitFile(_ string) (pulse.GitFile, error) {
	return f.file, nil
}

func (f *FileReader) SetFile(file pulse.GitFile) {
	f.file = file
}

func NewFileReader() *FileReader {
	return &FileReader{}
}
