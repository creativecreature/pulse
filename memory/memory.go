package memory

import (
	"github.com/creativecreature/pulse"
)

type Storage struct {
	sessions pulse.Sessions
}

func NewStorage() *Storage {
	return &Storage{sessions: pulse.Sessions{}}
}

func (m *Storage) Write(s pulse.Session) error {
	m.sessions = append(m.sessions, s)
	return nil
}

func (m *Storage) Read() (pulse.Sessions, error) {
	return m.sessions, nil
}

func (m *Storage) Clean() error {
	m.sessions = make(pulse.Sessions, 0)
	return nil
}
