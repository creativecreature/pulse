package main

import (
	"os"

	"code-harvest.conner.dev/internal/aggregate"
	"code-harvest.conner.dev/internal/storage"
	"code-harvest.conner.dev/pkg/logger"
)

// ldflags
var (
	uri string
	db  string
)

func main() {
	log := logger.New(os.Stdout, logger.LevelInfo)

	diskStorage := storage.DiskStorage()
	permStorage, disconnect := storage.MongoStorage(uri, db, "sessions")
	defer disconnect()

	err := aggregate.Day(diskStorage, permStorage)
	if err != nil {
		log.PrintFatal(err, nil)
	}

	log.PrintInfo("All temporary sessions were aggregated successfully", nil)
}
