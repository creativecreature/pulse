package memory

import (
	"github.com/creativecreature/code-harvest"
)

type Storage struct {
	sessions codeharvest.Sessions
}

func NewStorage() *Storage {
	return &Storage{sessions: codeharvest.Sessions{}}
}

func (m *Storage) Write(s codeharvest.Session) error {
	m.sessions = append(m.sessions, s)
	return nil
}

func (m *Storage) Read() (codeharvest.Sessions, error) {
	return m.sessions, nil
}

func (m *Storage) Clean() error {
	m.sessions = make(codeharvest.Sessions, 0)
	return nil
}
