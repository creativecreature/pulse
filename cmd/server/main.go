package main

import (
	"os"

	"github.com/creativecreature/pulse/disk"
	"github.com/creativecreature/pulse/logger"
	"github.com/creativecreature/pulse/server"
)

// ldflags.
var (
	serverName string
	port       string
)

func main() {
	log := logger.New(os.Stdout, logger.LevelDebug)
	diskStorage, err := disk.NewStorage()
	if err != nil {
		log.PrintFatal(err, nil)
	}

	server, err := server.New(
		serverName,
		server.WithLog(log),
		server.WithStorage(diskStorage),
	)
	if err != nil {
		log.PrintFatal(err, nil)
	}

	err = server.Start(port)
	if err != nil {
		log.PrintFatal(err, nil)
	}
}
