package disk

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/creativecreature/pulse"
)

const (
	YYYYMMDD  = "2006-01-02"
	HHMMSSSSS = "15:04:05.000"
)

// Storage implements the pulse.TemporaryStorage interface, and is used to store coding sessions on disk.
type Storage struct {
	root string
}

// NewStorage is used to create a new disk storage. It returns an error if it fails to create a
// "$HOME/.pulse" directory. Coding sessions will be written to within "$HOME/.pulse/tmp".
func NewStorage() (*Storage, error) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	// Create a .pulse directory in the home directory if it doesn't exist.
	root := path.Join(userHomeDir, ".pulse")
	if _, statErr := os.Stat(root); os.IsNotExist(statErr) {
		if mkdirErr := os.MkdirAll(root, os.ModePerm); mkdirErr != nil {
			return nil, mkdirErr
		}
	}

	return &Storage{root}, nil
}

// dayDir creates the directory where we'll store all coding sessions for a given day.
func (s *Storage) dayDir() (string, error) {
	dirPath := path.Join(s.root, "tmp", time.Now().UTC().Format(YYYYMMDD))
	// os.MkdirAll returns nil if the directory already exists
	err := os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		return "", err
	}

	return dirPath, nil
}

// filename returns the name we'll use when writing the session to disk.
func (s *Storage) filename(session pulse.Session) string {
	startTime := time.UnixMilli(session.StartedAt).Format("15:04:05.000")
	endTime := time.UnixMilli(session.EndedAt).Format("15:04:05.000")
	return fmt.Sprintf("%s-%s.json", startTime, endTime)
}

// Root returns the root directory for the storage.
func (s *Storage) Root() string {
	return s.root
}

// Write is used to write a coding session to disk.
func (s *Storage) Write(session pulse.Session) error {
	fName := s.filename(session)
	dDir, err := s.dayDir()
	if err != nil {
		return err
	}

	file, err := os.Create(path.Join(dDir, fName))
	if err != nil {
		return err
	}
	defer file.Close()

	bytes, err := session.Serialize()
	if err != nil {
		return err
	}
	_, err = file.Write(bytes)

	return err
}

// Read is used to read all coding sessions from disk.
func (s *Storage) Read() (pulse.Sessions, error) {
	temporarySessions := make(pulse.Sessions, 0)
	tmpDir := path.Join(s.root, "tmp")
	err := fs.WalkDir(os.DirFS(tmpDir), ".", func(p string, _ fs.DirEntry, _ error) error {
		if filepath.Ext(p) == ".json" {
			content, err := os.ReadFile(path.Join(tmpDir, p))
			if err != nil {
				return err
			}

			tempSession := pulse.Session{}
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

func (s *Storage) Clean() error {
	tmpDir := path.Join(s.root, "tmp")
	return os.RemoveAll(tmpDir)
}
