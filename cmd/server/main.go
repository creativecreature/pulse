package main

import (
	"context"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/viccon/pulse"
	"github.com/viccon/pulse/redis"
	"github.com/viccon/pulse/server"
)

func main() {
	cfg, err := pulse.ParseConfig()
	if err != nil {
		panic("failed to parse config")
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	redisClient := redis.New(cfg.Database.Address, cfg.Database.Password)

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	// Create the path for the log storages segment files.
	segmentPath := path.Join(userHomeDir, ".pulse", "segments")

	server := server.New(cfg, segmentPath, redisClient)
	server.RunBackgroundJobs(ctx, cfg.Server.SegmentationInterval)

	err = server.StartServer(ctx, cfg.Server.Port)
	if err != nil {
		panic(err)
	}
}
