// Package filestorage implements functions for temporarily storing our coding
// sessions to disk. The coding sessions are stored in the ~/.code-harvest/tmp
// directory. Each file in that directory is then being read by a cron job that
// transforms the data into a more suitable format. That data is then being
// saved in a database and served by our API.
package filestorage

import (
	"fmt"
	"os"
	"path"
	"time"

	"code-harvest.conner.dev/internal/domain"
	"code-harvest.conner.dev/internal/storage/data"
)

const (
	YYYYMMDD  = "2006-01-02"
	HHMMSSSSS = "15:04:05.000"
)

type Storage struct {
	dataDirPath string
}

func New(dataDirPath string) Storage {
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

	serializedSession, err := data.NewTemporarySession(domainSession).Serialize()
	if err != nil {
		return err
	}

	_, err = file.Write(serializedSession)
	return err
}
