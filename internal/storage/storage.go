package storage

import (
	"code-harvest.conner.dev/internal/domain"
	"code-harvest.conner.dev/internal/storage/mongodb"
)

type Storage interface {
	Connect() func()
	Save(s domain.Session) error
}

func MongoDB(uri, database, collection string) Storage {
	return mongodb.New(uri, database, collection)
}
