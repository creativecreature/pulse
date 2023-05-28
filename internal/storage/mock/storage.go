package mock

import (
	"code-harvest.conner.dev/internal/domain"
)

type Storage struct {
	sessions []domain.Session
}

func (m *Storage) Connect() func() {
	return func() {}
}

func (m *Storage) Save(s domain.Session) error {
	m.sessions = append(m.sessions, s)
	return nil
}

func (m *Storage) Get() []domain.Session {
	return m.sessions
}
