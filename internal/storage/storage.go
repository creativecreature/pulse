package storage

import (
	"code-harvest.conner.dev/internal/domain"
	"code-harvest.conner.dev/internal/storage/disk"
	"code-harvest.conner.dev/internal/storage/memory"
	"code-harvest.conner.dev/internal/storage/models"
	"code-harvest.conner.dev/internal/storage/mongo"
)

type TemporaryStorage interface {
	Save(s domain.Session) error
	GetAll() ([]models.TemporarySession, error)
	RemoveAll() error
}

func DiskStorage(dataDirPath string) TemporaryStorage {
	return disk.NewStorage(dataDirPath)
}

func MemoryStorage() TemporaryStorage {
	return memory.NewStorage()
}

type PermanentStorage interface {
	Save(s models.TemporarySession) error
}

func MongoStorage(uri, database, collection string) PermanentStorage {
	storage := mongo.NewDB(uri, database, collection)
	disconnect := storage.Connect()
	defer disconnect()
	return storage
}
