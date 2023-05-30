package memory

import (
	"code-harvest.conner.dev/internal/domain"
	"code-harvest.conner.dev/internal/storage/models"
)

type Storage struct {
	sessions []models.TemporarySession
}

func NewStorage() *Storage {
	return &Storage{sessions: []models.TemporarySession{}}
}

func (m *Storage) Save(s domain.Session) error {
	m.sessions = append(m.sessions, models.NewTemporarySession(s))
	return nil
}

func (m *Storage) GetAll() ([]models.TemporarySession, error) {
	return m.sessions, nil
}
