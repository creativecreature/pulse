package memory

import (
	"code-harvest.conner.dev/domain"
)

type Storage struct {
	sessions domain.Sessions
}

func NewStorage() *Storage {
	return &Storage{sessions: domain.Sessions{}}
}

func (m *Storage) Write(s domain.Session) error {
	m.sessions = append(m.sessions, s)
	return nil
}

func (m *Storage) Read() (domain.Sessions, error) {
	return m.sessions, nil
}

func (m *Storage) Clean() error {
	m.sessions = make(domain.Sessions, 0)
	return nil
}
