package main

import (
	"os"

	"code-harvest.conner.dev/internal/app"
	"code-harvest.conner.dev/internal/db"
	"code-harvest.conner.dev/pkg/logger"
)

// Set by linker flags
var uri string
var port string

func main() {
	application, err := app.New(
		app.WithLog(logger.New(os.Stdout, logger.LevelInfo)),
		app.WithStorage(db.New(uri, "codeharvest", "sessions")),
	)
	if err != nil {
		panic(err)
	}

	application.Start(port)
}
