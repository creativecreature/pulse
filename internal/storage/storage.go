package storage

import (
	"code-harvest.conner.dev/internal/domain"
	"code-harvest.conner.dev/internal/storage/disk"
	"code-harvest.conner.dev/internal/storage/memory"
	"code-harvest.conner.dev/internal/storage/mongo"
)

type TemporaryStorage interface {
	Save(s domain.ActiveSession) error
	GetAll() (domain.Sessions, error)
	RemoveAll() error
}

func DiskStorage() TemporaryStorage {
	return disk.NewStorage()
}

func MemoryStorage() TemporaryStorage {
	return memory.NewStorage()
}

type PermanentStorage interface {
	SaveAll(s []domain.AggregatedSession) error
}

func MongoStorage(uri, database, collection string) (mongoStorage PermanentStorage, disconnect func()) {
	storage := mongo.NewDB(uri, database, collection)
	disconnect = storage.Connect()
	return storage, disconnect
}
