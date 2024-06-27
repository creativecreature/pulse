package server

import (
	"github.com/charmbracelet/log"
	"github.com/creativecreature/pulse"
	"github.com/creativecreature/pulse/clock"
)

type Option func(*Server)

// WithClock sets the clock used by the server.
func WithClock(clock clock.Clock) Option {
	return func(a *Server) {
		a.clock = clock
	}
}

// FileReader is a simple abstraction that defines a function
// for getting metadata from a file within a git repository.
type FileReader interface {
	GitFile(path string) (pulse.GitFile, error)
}

// WithLog sets the logger used by the server.
func WithLog(log *log.Logger) Option {
	return func(a *Server) {
		a.log = log
	}
}
