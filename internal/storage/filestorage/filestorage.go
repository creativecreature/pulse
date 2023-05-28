package filestorage

import (
	"fmt"
	"os"

	"code-harvest.conner.dev/internal/domain"
)

type Storage struct {
	dataDirPath string
}

func New(dataDirPath string) Storage {
	return Storage{dataDirPath}
}

func (s Storage) Save(domainSession domain.Session) error {
	filepath, err := s.createFile()
	if err != nil {
		return err
	}

	// Open the file in append mode
	file, err := os.OpenFile(filepath, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	serializedSession, err := newSession(domainSession).serialize()
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(file, "%s\n", serializedSession)
	return err
}
