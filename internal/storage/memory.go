package storage

import (
	"code-harvest.conner.dev/internal/models"
)

type MemoryStorage struct {
	sessions []*models.Session
}

func (m *MemoryStorage) Save(s *models.Session) error {
	m.sessions = append(m.sessions, s)
	return nil
}

func (m *MemoryStorage) Get() []*models.Session {
	return m.sessions
}
