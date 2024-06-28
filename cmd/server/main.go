package main

import (
	"context"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/creativecreature/pulse"
	"github.com/creativecreature/pulse/mongo"
	"github.com/creativecreature/pulse/server"
)

func main() {
	cfg, err := pulse.ParseConfig()
	if err != nil {
		panic("failed to parse config")
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	client := mongo.New(cfg.Database.URI, cfg.Database.Name)
	defer func() {
		disconnectContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		disconnectErr := client.Disconnect(disconnectContext)
		if disconnectErr != nil {
			panic(disconnectErr)
		}
	}()

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	// Create the path for the log storages segment files.
	segmentPath := path.Join(userHomeDir, ".pulse", "segments")

	server := server.New(cfg, segmentPath, client)
	server.RunBackgroundJobs(ctx, cfg.Server.SegmentationInterval)

	err = server.StartServer(ctx, cfg.Server.Port)
	if err != nil {
		panic(err)
	}
}
