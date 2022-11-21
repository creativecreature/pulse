package app

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

func (app *app) CheckHeartbeat() {
	app.log.PrintDebug("Checking heartbeat", nil)
	if app.session != nil && app.lastHeartbeat+HeartbeatTTL.Milliseconds() < app.clock.GetTime() {
		app.mutex.Lock()
		defer app.mutex.Unlock()
		app.saveSession()
	}
}

// Listen for shutdown signals and perform heartbeat checks.
func (app *app) monitorHeartbeat() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	ecg := time.NewTicker(heartbeatInterval)

	run := true
	for run {
		select {
		case <-ecg.C:
			app.CheckHeartbeat()
		case <-quit:
			run = false
		}
	}

	ecg.Stop()
}
