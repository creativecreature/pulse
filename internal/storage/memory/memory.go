package memory

import (
	"code-harvest.conner.dev/internal/domain"
)

type Storage struct {
	sessions domain.Sessions
}

func NewStorage() *Storage {
	return &Storage{sessions: domain.Sessions{}}
}

func (m *Storage) Save(s domain.ActiveSession) error {
	m.sessions = append(m.sessions, domain.NewSession(s))
	return nil
}

func (m *Storage) GetAll() (domain.Sessions, error) {
	return m.sessions, nil
}

func (m *Storage) RemoveAll() error {
	m.sessions = make(domain.Sessions, 0)
	return nil
}
