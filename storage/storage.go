package storage

import (
	"code-harvest.conner.dev/domain"
	"code-harvest.conner.dev/storage/disk"
	"code-harvest.conner.dev/storage/memory"
	"code-harvest.conner.dev/storage/mongo"
)

type TemporaryStorage interface {
	Write(domain.Session) error
	Read() (domain.Sessions, error)
	Clean() error
}

func DiskStorage() TemporaryStorage {
	return disk.NewStorage()
}

func MemoryStorage() TemporaryStorage {
	return memory.NewStorage()
}

type PermanentStorage interface {
	Write(s []domain.AggregatedSession) error
	Aggregate(timeperiod domain.TimePeriod) error
}

func MongoStorage(uri, database string) (mongoStorage PermanentStorage, disconnect func()) {
	storage := mongo.NewDB(uri, database)
	disconnect = storage.Connect()
	return storage, disconnect
}
