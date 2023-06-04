// Package disk implements functions for temporarily storing our coding
// sessions to disk. The coding sessions are stored in the ~/.code-harvest/tmp
// directory. Each file in that directory is then being read by a cron job that
// transforms the data into a more suitable format. That data is then being
// saved in a database and served by our API.
package disk

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"time"

	"code-harvest.conner.dev/internal/domain"
	"code-harvest.conner.dev/internal/storage/models"
)

const (
	YYYYMMDD  = "2006-01-02"
	HHMMSSSSS = "15:04:05.000"
)

type Storage struct {
	dataDirPath string
}

func NewStorage() Storage {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	dataDirPath := path.Join(homeDir, ".code-harvest")
	return Storage{dataDirPath}
}

// dir creates the directory where we'll store all coding sessions for a given day
func dir(dataDirPath string) (string, error) {
	dirPath := path.Join(dataDirPath, "tmp", time.Now().UTC().Format(YYYYMMDD))
	// os.MkdirAll returns nil if the directory already exists
	err := os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		return "", err
	}

	return dirPath, nil
}

// Returns a filename that we'll use when writing the session to disk
func filename(s domain.Session) string {
	startDuration := time.Duration(s.StartedAt) * time.Millisecond
	startTime := time.Unix(0, startDuration.Nanoseconds())
	endDuration := time.Duration(s.EndedAt) * time.Millisecond
	endTime := time.Unix(0, endDuration.Nanoseconds())
	return fmt.Sprintf("%s-%s.json", startTime.Format(HHMMSSSSS), endTime.Format(HHMMSSSSS))
}

func (s Storage) Save(domainSession domain.Session) error {
	fname := filename(domainSession)
	dirPath, err := dir(s.dataDirPath)
	if err != nil {
		return err
	}

	file, err := os.Create(path.Join(dirPath, fname))
	if err != nil {
		return err
	}
	defer file.Close()

	serializedSession, err := models.NewTemporarySession(domainSession).Serialize()
	if err != nil {
		return err
	}

	_, err = file.Write(serializedSession)
	return err
}

func (s Storage) GetAll() ([]models.TemporarySession, error) {
	temporarySessions := make([]models.TemporarySession, 0)
	tmpDir := path.Join(s.dataDirPath, "tmp")
	fmt.Println(tmpDir)
	err := fs.WalkDir(os.DirFS(tmpDir), ".", func(p string, _ fs.DirEntry, _ error) error {
		if filepath.Ext(p) == ".json" {
			content, err := os.ReadFile(path.Join(tmpDir, p))
			if err != nil {
				return err
			}
			tempSession := models.TemporarySession{}
			err = json.Unmarshal(content, &tempSession)
			if err != nil {
				return err
			}
			temporarySessions = append(temporarySessions, tempSession)
		}
		return nil
	})

	return temporarySessions, err
}

func (s Storage) RemoveAll() error {
	tmpDir := path.Join(s.dataDirPath, "tmp")
	return os.RemoveAll(tmpDir)
}
