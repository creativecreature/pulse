package filestorage

import (
	"os"
	"path"
	"time"
)

const (
	YYYYMMDD = "2006-01-02"
)

func (s Storage) filepath() string {
	return path.Join(s.dataDirPath, time.Now().UTC().Format(YYYYMMDD))
}

// createFile makes sure that the file that we're going to save the session to
// exists. If it doesn't exist, it creates it.
func (s Storage) createFile() (string, error) {
	filepath := s.filepath()
	_, err := os.Stat(filepath)
	if os.IsNotExist(err) {
		_, err := os.Create(filepath)
		if err != nil {
			return "", err
		}
	} else if err != nil {
		return "", err
	}
	return filepath, nil
}
