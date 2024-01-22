package main

import (
	"os"

	"github.com/creativecreature/code-harvest/disk"
	"github.com/creativecreature/code-harvest/logger"
	"github.com/creativecreature/code-harvest/server"
)

// ldflags.
var (
	serverName string
	port       string
)

func main() {
	log := logger.New(os.Stdout, logger.LevelDebug)

	server, err := server.New(
		serverName,
		server.WithLog(log),
		server.WithStorage(disk.NewStorage()),
	)
	if err != nil {
		log.PrintFatal(err, nil)
	}

	err = server.Start(port)
	if err != nil {
		log.PrintFatal(err, nil)
	}
}
