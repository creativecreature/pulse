package memory

import (
	"code-harvest.conner.dev/internal/domain"
)

type Storage struct {
	sessions []domain.Session
}

func NewStorage() *Storage {
	return &Storage{sessions: []domain.Session{}}
}

func (m *Storage) Save(s domain.ActiveSession) error {
	m.sessions = append(m.sessions, domain.NewSession(s))
	return nil
}

func (m *Storage) GetAll() ([]domain.Session, error) {
	return m.sessions, nil
}

func (m *Storage) RemoveAll() error {
	m.sessions = make([]domain.Session, 0)
	return nil
}
