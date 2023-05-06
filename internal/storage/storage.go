package storage

import (
	"code-harvest.conner.dev/internal/storage/mongodb"
)

type Storage interface {
	Connect() func()
	Save(s interface{}) error
}

func MongoDB(uri, database, collection string) Storage {
	return mongodb.New(uri, database, collection)
}
