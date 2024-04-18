package main

import (
	"context"
	"time"

	"github.com/creativecreature/pulse/disk"
	"github.com/creativecreature/pulse/logger"
	"github.com/creativecreature/pulse/mongo"
)

// ldflags.
var (
	uri string
	db  string
)

func main() {
	log := logger.New()
	diskStorage, err := disk.NewStorage()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	timeoutCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	client := mongo.New(uri, db, log)
	defer func() {
		disconnectErr := client.Disconnect(timeoutCtx)
		if disconnectErr != nil {
			log.Fatal(disconnectErr)
		}
	}()

	log.Info("Reading sessions from disk.")
	diskSessions, err := diskStorage.Read()
	if err != nil {
		log.Error(err)
		return
	}

	if len(diskSessions) < 1 {
		log.Info("Found no sessions to aggregate.")
		return
	}

	log.Infof("Found %d sessions. Initiating database writes.", len(diskSessions))
	err = client.Write(timeoutCtx, diskSessions.Aggregate())
	if err != nil {
		log.Error(err)
		return
	}

	log.Info("The sessions have been aggregated successfully. Removing them from disk.")
	err = diskStorage.Clean()
	if err != nil {
		log.Error(err)
		return
	}
	log.Info("Finished the aggregation.")
}
