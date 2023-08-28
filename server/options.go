package server

import (
	"errors"

	"github.com/creativecreature/code-harvest"
	"github.com/creativecreature/code-harvest/clock"
	"github.com/creativecreature/code-harvest/filereader"
	"github.com/creativecreature/code-harvest/storage"
)

type option func(*server) error

// Clock is a simple abstraction that is used to
// simplify time based assertions in tests.
type Clock interface {
	GetTime() int64
}

// WithClock sets the clock used by the server.
func WithClock(clock Clock) option {
	return func(a *server) error {
		if clock == nil {
			return errors.New("clock is nil")
		}
		a.clock = clock
		return nil
	}
}

// FileReader is a simple abstraction that defines a function
// for getting metadata from a file within a git repository.
type FileReader interface {
	GitFile(path string) (codeharvest.GitFile, error)
}

// WithFileReader sets the file reader used by the server.
func WithFileReader(reader FileReader) option {
	return func(a *server) error {
		if reader == nil {
			return errors.New("reader is nil")
		}
		a.fileReader = reader
		return nil
	}
}

// WithStorage sets the storage used by the server.
func WithStorage(storage storage.TemporaryStorage) option {
	return func(a *server) error {
		if storage == nil {
			return errors.New("storage is nil")
		}
		a.storage = storage
		return nil
	}
}

// Log is an abstraction for the logger used by the server.
// It allows us to use a different logger during tests.
type Log interface {
	PrintDebug(message string, properties map[string]string)
	PrintInfo(message string, properties map[string]string)
	PrintError(err error, properties map[string]string)
	PrintFatal(err error, properties map[string]string)
}

// WithLog sets the logger used by the server.
func WithLog(log Log) option {
	return func(a *server) error {
		if log == nil {
			return errors.New("log is nil")
		}
		a.log = log
		return nil
	}
}

// New creates a new server.
func New(serverName string, opts ...option) (*server, error) {
	a := &server{
		serverName: serverName,
		clock:      clock.New(),
		fileReader: filereader.New(),
	}
	for _, opt := range opts {
		err := opt(a)
		if err != nil {
			return &server{}, err
		}
	}
	return a, nil
}
