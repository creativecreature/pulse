package storage

import (
	"code-harvest.conner.dev/internal/models"
)

type Storage interface {
	Save(s *models.Session) error
}
