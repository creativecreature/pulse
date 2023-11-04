package mock

import (
	codeharvest "github.com/creativecreature/code-harvest"
)

// FileReader is a mock implementation of the domain.FileReader interface.
type FileReader struct {
	file codeharvest.GitFile
}

func (f *FileReader) GitFile(_ string) (codeharvest.GitFile, error) {
	return f.file, nil
}

func (f *FileReader) SetFile(file codeharvest.GitFile) {
	f.file = file
}

func NewFileReader() *FileReader {
	return &FileReader{}
}
