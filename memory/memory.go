package memory

import (
	"sync"

	"github.com/creativecreature/pulse"
)

type Storage struct {
	sync.Mutex
	sessions pulse.Sessions
}

func NewStorage() *Storage {
	return &Storage{sessions: pulse.Sessions{}}
}

func (m *Storage) Write(s pulse.Session) error {
	m.Lock()
	defer m.Unlock()
	m.sessions = append(m.sessions, s)
	return nil
}

func (m *Storage) Read() (pulse.Sessions, error) {
	m.Lock()
	defer m.Unlock()
	return m.sessions, nil
}

func (m *Storage) Clean() error {
	m.Lock()
	defer m.Unlock()
	m.sessions = make(pulse.Sessions, 0)
	return nil
}
