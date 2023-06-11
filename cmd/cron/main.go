package main

import (
	"os"

	"code-harvest.conner.dev/internal/storage"
	"code-harvest.conner.dev/pkg/logger"
)

// Set by linker flags
var (
	uri        string
	db         string
	collection string
)

func main() {
	log := logger.New(os.Stdout, logger.LevelInfo)
	diskStorage := storage.DiskStorage()
	sessions, err := diskStorage.GetAll()
	if err != nil {
		log.PrintFatal(err, nil)
	}

	permStorage, disconnect := storage.MongoStorage(uri, db, collection)
	defer disconnect()
	err = permStorage.SaveAll(sessions.AggregateByDay())
	if err != nil {
		log.PrintFatal(err, nil)
	}

	err = diskStorage.RemoveAll()
	if err != nil {
		log.PrintFatal(err, nil)
	}
}
