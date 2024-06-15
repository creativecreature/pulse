package server

import (
	"errors"

	"github.com/charmbracelet/log"
	"github.com/creativecreature/pulse"
)

type Option func(*Server) error

// WithClock sets the clock used by the server.
func WithClock(clock Clock) Option {
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

// WithLog sets the logger used by the server.
func WithLog(log *log.Logger) Option {
	return func(a *Server) error {
		if log == nil {
			return errors.New("log is nil")
		}
		a.log = log
		return nil
	}
}
