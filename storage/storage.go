package storage

import (
	codeharvest "github.com/creativecreature/code-harvest"
	"github.com/creativecreature/code-harvest/storage/disk"
	"github.com/creativecreature/code-harvest/storage/memory"
	"github.com/creativecreature/code-harvest/storage/mongo"
)

type TemporaryStorage interface {
	Write(codeharvest.Session) error
	Read() (codeharvest.Sessions, error)
	Clean() error
}

func DiskStorage() TemporaryStorage {
	return disk.NewStorage()
}

func MemoryStorage() TemporaryStorage {
	return memory.NewStorage()
}

type PermanentStorage interface {
	Write(s []codeharvest.AggregatedSession) error
	Aggregate(timeperiod codeharvest.TimePeriod) error
}

func MongoStorage(uri, database string) (mongoStorage PermanentStorage, disconnect func()) {
	storage := mongo.NewDB(uri, database)
	disconnect = storage.Connect()
	return storage, disconnect
}
