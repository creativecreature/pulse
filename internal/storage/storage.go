package storage

import (
	"code-harvest.conner.dev/internal/domain"
	"code-harvest.conner.dev/internal/storage/data"
	"code-harvest.conner.dev/internal/storage/filestorage"
	"code-harvest.conner.dev/internal/storage/mongodb"
)

type TemporaryStorage interface {
	Save(s domain.Session) error
}

func Filestorage(dataDirPath string) TemporaryStorage {
	return filestorage.New(dataDirPath)
}

type PermanentStorage interface {
	Save(s data.TemporarySession) error
}

func MongoDB(uri, database, collection string) PermanentStorage {
	storage := mongodb.New(uri, database, collection)
	disconnect := storage.Connect()
	defer disconnect()
	return storage
}
