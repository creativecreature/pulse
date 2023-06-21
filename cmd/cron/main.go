package main

import (
	"os"

	"code-harvest.conner.dev/domain"
	"code-harvest.conner.dev/logger"
	"code-harvest.conner.dev/storage"
)

// ldflags
var (
	uri string
	db  string
)

func aggregateTempSessions(tempStorage storage.TemporaryStorage, permStorage storage.PermanentStorage) error {
	tempSessions, err := tempStorage.Read()
	if err != nil {
		return err
	}
	err = permStorage.Write(tempSessions.Aggregate())
	if err != nil {
		return err
	}
	return tempStorage.Clean()
}

func main() {
	log := logger.New(os.Stdout, logger.LevelInfo)

	diskStorage := storage.DiskStorage()
	permStorage, disconnect := storage.MongoStorage(uri, db)
	defer disconnect()

	err := aggregateTempSessions(diskStorage, permStorage)
	if err != nil {
		log.PrintFatal(err, nil)
	}

	log.PrintInfo("All temporary sessions were aggregated successfully", nil)

	err = permStorage.Aggregate(domain.Week)
	if err != nil {
		log.PrintFatal(err, nil)
	}
	log.PrintInfo("All weekly sessions were aggregated successfully", nil)

	err = permStorage.Aggregate(domain.Month)
	if err != nil {
		log.PrintFatal(err, nil)
	}
	log.PrintInfo("All monthly sessions were aggregated successfully", nil)

	err = permStorage.Aggregate(domain.Year)
	if err != nil {
		log.PrintFatal(err, nil)
	}
	log.PrintInfo("All yearly sessions were aggregated successfully", nil)
}
