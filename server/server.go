package server

import (
	"errors"
	"fmt"
	"math"
	"net"
	"net/http"
	"net/rpc"
	"sync"

	"code-harvest.conner.dev/domain"
	"code-harvest.conner.dev/proxy"
	"code-harvest.conner.dev/storage"
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
	server.session = domain.StartSession(server.clock.GetTime(), os, editor)
}

// setActiveBuffer updates the current buffer in the current session.
func (server *server) setActiveBuffer(absolutePath string) {
	openedAt := server.clock.GetTime()
	fileData, err := server.fileReader.GitFile(absolutePath)
	if err != nil {
		server.log.PrintDebug("Could not extract metadata for the path", map[string]string{
			"path":   absolutePath,
			"reason": err.Error(),
		})
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
	server.log.PrintDebug("Successfully updated the current buffer", map[string]string{
		"path": absolutePath,
	})
}

// hasOkDurations sanity checks the sessions total duration against
// the combined duration of all files that were opened. Sometimes
// I edit files that I don't want to track but it should not
// differ by more than 25%.
func hasOkDurations(sessionDuration, allFilesDuration int64) bool {
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

// Start starts the server on the given port.
func (server *server) Start(port string) error {
	server.log.PrintInfo("Starting up...", nil)

	// The proxy exposes the functions that we want to make available for
	// remote procedure calls. We then register the proxy as the RPC receiver.
	proxy := proxy.New(server)
	err := rpc.RegisterName(server.serverName, proxy)
	if err != nil {
		return err
	}

	rpc.HandleHTTP()
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		return err
	}

	err = http.Serve(listener, nil)
	if err != nil {
		return err
	}
	defer listener.Close()

	// Blocks until we receive a shutdown signal
	server.monitorHeartbeat()

	// Save the current session before shutting down
	server.saveSession()

	server.log.PrintInfo("Shutting down...", nil)
	return nil
}
