package server

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"code-harvest.conner.dev/domain"
	"code-harvest.conner.dev/proxy"
	"code-harvest.conner.dev/storage"
)

const (
	HeartbeatTTL      = time.Minute * 10
	heartbeatInterval = time.Second * 45
)

type server struct {
	activeClientId string
	clock          Clock
	lastHeartbeat  int64
	log            Log
	fileReader     FileReader
	mutex          sync.Mutex
	serverName     string
	session        *domain.ActiveSession
	storage        storage.TemporaryStorage
}

// startNewSession creates a new session and sets it as the current session.
func (server *server) startNewSession(os, editor string) {
	server.log.PrintDebug("Starting a new session", nil)
	server.session = domain.StartSession(server.clock.GetTime(), os, editor)
}

// setActiveBuffer updates the current buffer in the current session.
func (server *server) setActiveBuffer(absolutePath string) {
	openedAt := server.clock.GetTime()
	fileData, err := server.fileReader.GitFile(absolutePath)
	if err != nil {
		exitEarlyMessage := "Could not extract metadata for the path"
		errProperties := map[string]string{
			"path":   absolutePath,
			"reason": err.Error(),
		}
		server.log.PrintDebug(exitEarlyMessage, errProperties)
		return
	}

	file := domain.NewBuffer(
		fileData.Name,
		fileData.Repository,
		fileData.Filetype,
		fileData.Path,
		openedAt,
	)

	server.session.PushBuffer(file)
	updatedBufferMsg := "Successfully updated the current buffer"
	updatedBufferProperties := map[string]string{
		"path": absolutePath,
	}
	server.log.PrintDebug(updatedBufferMsg, updatedBufferProperties)
}

// hasOkDurations sanity checks the sessions total duration against
// the combined duration of all files that were opened.
func hasOkDurations(sessionDuration, allFilesDuration int64) bool {
	// Exclude sessions that are less than 10 minutes.
	tenMinutesMS := int64(10 * 60 * 1000)
	if sessionDuration < tenMinutesMS {
		return true
	}
	// If the session lasts for more than 10 minutes, and the time
	// differs by more than 25%, we'll want to check the session.
	larger := math.Max(float64(sessionDuration), float64(allFilesDuration))
	threshold := larger * 0.25
	difference := math.Abs(float64(sessionDuration) - float64(allFilesDuration))
	return difference <= threshold
}

// saveSession ends the current coding session and saves it to the filesystem.
func (server *server) saveSession() {
	if server.session == nil {
		server.log.PrintDebug("There was no session to save.", nil)
		return
	}

	// End the current session.
	server.log.PrintDebug("Saving the session.", nil)
	endedAt := server.clock.GetTime()
	finishedSession := server.session.End(endedAt)

	// Perform sanity checks on the durations.
	totalFileDuration := finishedSession.TotalFileDuration()
	if !hasOkDurations(finishedSession.DurationMs, totalFileDuration) {
		finishedSession.DurationMs = totalFileDuration
		error := errors.New("session had a large duration diff")
		properties := map[string]string{
			"started_at":                fmt.Sprintf("%d", server.session.StartedAt),
			"ended_at":                  fmt.Sprintf("%d", endedAt),
			"previous_session_duration": fmt.Sprintf("%d", finishedSession.DurationMs),
			"new_session_duration":      fmt.Sprintf("%d", totalFileDuration),
		}
		server.log.PrintError(error, properties)
	}

	err := server.storage.Write(finishedSession)
	if err != nil {
		server.log.PrintError(err, nil)
	}

	server.activeClientId = ""
	server.session = nil
}

func (s *server) startServer(port string) (*http.Server, error) {
	proxy := proxy.New(s)
	err := rpc.RegisterName(s.serverName, proxy)
	if err != nil {
		return nil, err
	}

	rpc.HandleHTTP()
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		return nil, err
	}
	httpServer := &http.Server{}
	go func() {
		err := httpServer.Serve(listener)
		if !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	return httpServer, nil
}

// startECG runs a heartbeat ticker that ensures that
// the current session is not idle for more than ten minutes.
func (server *server) startECG() *time.Ticker {
	server.log.PrintDebug("Starting the ECG", nil)
	ecg := time.NewTicker(heartbeatInterval)
	go func() {
		for range ecg.C {
			server.CheckHeartbeat()
		}
	}()
	return ecg
}

// Start starts the server on the given port.
func (server *server) Start(port string) error {
	server.log.PrintInfo("Starting up...", nil)
	httpServer, err := server.startServer(port)
	if err != nil {
		return err
	}

	// Start the ECG. It will end inactive sessions.
	ecg := server.startECG()

	// Catch shutdown signals from the OS
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Blocks until a shutdown signal is received.
	s := <-quit
	server.log.PrintInfo("Received shutdown signal", map[string]string{
		"signal": s.String(),
	})

	// Stop the ECG and shutdown the http server.
	ecg.Stop()
	err = httpServer.Shutdown(context.Background())
	if err != nil {
		server.log.PrintError(err, nil)
		return err
	}

	// Save the current session before shutting down.
	server.saveSession()
	server.log.PrintInfo("Shutting down...", nil)
	return nil
}
