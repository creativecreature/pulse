package server

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	HeartbeatTTL      = time.Minute * 10
	heartbeatInterval = time.Second * 10
)

// CheckHeartbeat is used to check if the session has been
// inactive for more than ten minutes. If that is the case,
// the session will be terminated and saved to disk.
func (server *server) CheckHeartbeat() {
	server.log.PrintDebug("Checking heartbeat", nil)
	if server.session == nil {
		return
	}

	if server.lastHeartbeat+HeartbeatTTL.Milliseconds() < server.clock.GetTime() {
		// Lock the mutex to prevent race conditions with events from the clients.
		server.mutex.Lock()
		defer server.mutex.Unlock()
		server.log.PrintDebug("Ending inactive session", map[string]string{
			"last_heartbeat": fmt.Sprintf("%d", server.lastHeartbeat),
			"current_time":   fmt.Sprintf("%d", server.clock.GetTime()),
		})
		server.saveSession()
	}
}

// monitorHeartbeat runs a heartbeat ticker that ensures that
// the current session is not idle for more than ten minutes.
func (server *server) monitorHeartbeat() {
	// We'll listen to SIGTERM and SIGINT to ensure that we
	// end the heartbeat before the server shuts down.
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
