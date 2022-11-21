package app

import (
	"sync"

	"code-harvest.conner.dev/internal/models"
	"code-harvest.conner.dev/pkg/clock"
	"code-harvest.conner.dev/pkg/logger"
)

type FileMetadata struct {
	Filename       string
	Filetype       string
	RepositoryName string
}

type FileMetadataReader interface {
	Read(uri string) (FileMetadata, error)
}

type storage interface {
	Connect() func()
	Save(s interface{}) error
}

type app struct {
	mutex          sync.Mutex
	clock          clock.Clock
	metadataReader FileMetadataReader
	storage        storage
	activeClientId string
	lastHeartbeat  int64
	session        *models.Session
	log            *logger.Logger
}

func (app *app) Run(port string) error {
	app.log.PrintInfo("Starting up...", nil)

	// Connect to the storage
	disconnect := app.storage.Connect()
	defer disconnect()

	// Start the RPC server
	listener, err := startServer(app, port)
	if err != nil {
		app.log.PrintFatal(err, nil)
	}

	// Blocks until we receive a shutdown signal
	app.monitorHeartbeat()

	app.log.PrintInfo("Shutting down...", nil)
	return listener.Close()
}
