package server

import (
	"errors"

	"github.com/creativecreature/pulse"
)

type Option func(*Server) error

// WithClock sets the clock used by the server.
func WithClock(clock pulse.Clock) Option {
	return func(a *Server) error {
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
	GitFile(path string) (pulse.GitFile, error)
}

// WithFileReader sets the file reader used by the server.
func WithFileReader(reader FileReader) Option {
	return func(a *Server) error {
		if reader == nil {
			return errors.New("reader is nil")
		}
		a.fileReader = reader
		return nil
	}
}

// WithStorage sets the storage used by the server.
func WithStorage(storage pulse.TemporaryStorage) Option {
	return func(a *Server) error {
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
func WithLog(log Log) Option {
	return func(a *Server) error {
		if log == nil {
			return errors.New("log is nil")
		}
		a.log = log
		return nil
	}
}
