package storage

import (
	"code-harvest.conner.dev/internal/domain"
	"code-harvest.conner.dev/internal/storage/filestorage"
	"code-harvest.conner.dev/internal/storage/mongodb"
)

type Storage interface {
	Save(s domain.Session) error
}

func MongoDB(uri, database, collection string) Storage {
	storage := mongodb.New(uri, database, collection)
	disconnect := storage.Connect()
	defer disconnect()
	return storage
}

func Filestorage(dataDirPath string) Storage {
	return filestorage.New(dataDirPath)
}
