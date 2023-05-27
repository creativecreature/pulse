package mock

import (
	"errors"

	"code-harvest.conner.dev/internal/domain"
)

type Storage struct {
	sessions []*domain.Session
}

func (m *Storage) Connect() func() {
	return func() {}
}

func (m *Storage) Save(s interface{}) error {
	result, ok := s.(*domain.Session)
	if !ok {
		return errors.New("failed to convert interface to slice of session pointers")
	}
	m.sessions = append(m.sessions, result)
	return nil
}

func (m *Storage) Get() []*domain.Session {
	return m.sessions
}
