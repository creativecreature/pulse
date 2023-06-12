package server

import (
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"sync"

	"code-harvest.conner.dev/internal/domain"
	"code-harvest.conner.dev/internal/proxy"
	"code-harvest.conner.dev/internal/storage"
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

func (server *server) startNewSession(os, editor string) {
	server.session = domain.NewActiveSession(server.clock.GetTime(), os, editor)
}

func (server *server) updateCurrentFile(absolutePath string) {
	openedAt := server.clock.GetTime()

	fileData, err := server.fileReader.GitFile(absolutePath)
	if err != nil {
		server.log.PrintDebug("Could not extract metadata for the path", map[string]string{
			"reason": err.Error(),
		})
		return
	}

	file := domain.NewBuffer(
		fileData.Name(),
		fileData.Repository(),
		fileData.Filetype(),
		fileData.Path(),
		openedAt,
	)

	// Update the current file.
	if currentBuffer := server.session.PeekBuffer(); currentBuffer != nil {
		currentBuffer.ClosedAt = openedAt
	}
	server.session.PushBuffer(file)
	server.log.PrintDebug("Successfully updated the current buffer", map[string]string{
		"path": absolutePath,
	})
}

func (server *server) saveSession() {
	// Regardless of how we exit this function we want to reset these values.
	defer func() {
		server.activeClientId = ""
		server.session = nil
	}()

	if server.session == nil {
		server.log.PrintDebug("There was no session to save.", nil)
		return
	}

	server.log.PrintDebug("Saving the session.", nil)

	// Set session duration and set closed at for the current file.
	endedAt := server.clock.GetTime()
	if currentFile := server.session.PeekBuffer(); currentFile != nil {
		currentFile.ClosedAt = endedAt
	}
	server.session.EndedAt = endedAt
	server.session.DurationMs = server.session.EndedAt - server.session.StartedAt

	// Whenever we open new a buffer that have a corresponding file on disk we
	// push it to the sessions file stack. Each buffer can be opened more than
	// once. Before we save the session we aggregate all the edits of the same
	// file into a map with a total duration of the time we've spent in that
	// file.
	for {
		buffer := server.session.PopBuffer()
		if buffer == nil {
			break
		}
		mergedBuffer, exists := server.session.MergedBuffers[buffer.Filepath]
		if !exists {
			buffer.DurationMs = buffer.ClosedAt - buffer.OpenedAt
			server.session.MergedBuffers[buffer.Filepath] = buffer
		} else {
			mergedBuffer.DurationMs += buffer.ClosedAt - buffer.OpenedAt
		}
	}

	if len(server.session.MergedBuffers) < 1 {
		server.log.PrintDebug("The session had no files.", map[string]string{
			"clientId": server.activeClientId,
		})
		return
	}

	err := server.storage.Save(*server.session)
	if err != nil {
		server.log.PrintError(err, nil)
	}
}

func (server *server) Start(port string) error {
	server.log.PrintInfo("Starting up...", nil)

	// The proxy exposes the functions that we want to make available for remote
	// procedure calls. Register the proxy as the RPC receiver.
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

	server.log.PrintInfo("Shutting down...", nil)
	return nil
}
