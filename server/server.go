package server

import (
	"fmt"
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

	// Set session duration and close the current buffer
	endedAt := server.clock.GetTime()
	endedSession := server.session.End(endedAt)

	err := server.storage.Write(endedSession)
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
