package server

import (
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	HeartbeatTTL      = time.Minute * 10
	heartbeatInterval = time.Second * 10
)

func (server *server) CheckHeartbeat() {
	server.log.PrintDebug("Checking heartbeat", nil)
	if server.session == nil {
		return
	}

	// Check if too much time has passed since the last heartbeat
	if server.lastHeartbeat+HeartbeatTTL.Milliseconds() < server.clock.GetTime() {
		server.mutex.Lock()
		defer server.mutex.Unlock()
		server.saveSession()
	}
}

// Listen for shutdown signals and perform heartbeat checks.
func (server *server) monitorHeartbeat() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	ecg := time.NewTicker(heartbeatInterval)

	run := true
	for run {
		select {
		case <-ecg.C:
			server.CheckHeartbeat()
		case <-quit:
			run = false
		}
	}

	ecg.Stop()
}
