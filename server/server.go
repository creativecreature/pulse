package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/charmbracelet/log"
	"github.com/creativecreature/pulse"
	"github.com/creativecreature/pulse/filereader"
	"github.com/creativecreature/pulse/proxy"
)

type Server struct {
	name                string
	activeEditorID      string
	activeSessions      map[string]*pulse.CodingSession
	lastHeartbeat       int64
	stopHeartbeatChecks chan struct{}
	clock               pulse.Clock
	fileReader          FileReader
	log                 *log.Logger
	mutex               sync.Mutex
	storage             pulse.TemporaryStorage
}

// New creates a new server.
func New(serverName string, opts ...Option) (*Server, error) {
	s := &Server{
		name:                serverName,
		activeSessions:      make(map[string]*pulse.CodingSession),
		clock:               pulse.NewClock(),
		stopHeartbeatChecks: make(chan struct{}),
		fileReader:          filereader.New(),
	}
	for _, opt := range opts {
		err := opt(s)
		if err != nil {
			return &Server{}, err
		}
	}
	s.startHeartbeatChecks()
	return s, nil
}

// createSession creates a new session and sets it as the current session.
func (s *Server) createSession(id, os, editor string) {
	s.log.Debug("Creating a new session.", "editor_id", id, "editor", editor, "os", os)
	s.activeSessions[id] = pulse.StartSession(id, s.clock.GetTime(), os, editor)
}

// setActiveBuffer updates the current buffer in the current session.
func (s *Server) setActiveBuffer(gitFile pulse.GitFile) {
	openedAt := s.clock.GetTime()
	buf := pulse.NewBuffer(
		gitFile.Name,
		gitFile.Repository,
		gitFile.Filetype,
		gitFile.Path,
		openedAt,
	)
	s.activeSessions[s.activeEditorID].PushBuffer(buf)
	s.log.Debug("Successfully updated the current buffer.",
		"name", gitFile.Name,
		"relative_path", gitFile.Path,
		"repository", gitFile.Repository,
		"filetype", gitFile.Filetype,
		"editor_id", s.activeEditorID,
		"editor", s.activeSessions[s.activeEditorID].Editor,
		"os", s.activeSessions[s.activeEditorID].OS,
	)
}

func (s *Server) saveAllSessions(endedAt int64) {
	s.log.Debug("Saving all sessions.")

	for _, session := range s.activeSessions {
		if !session.HasBuffers() {
			s.log.Debug("The session had not opened any buffers.",
				"editor_id", session.EditorID,
				"editor", session.Editor,
				"os", session.OS,
			)
			return
		}

		finishedSession := session.End(endedAt)
		err := s.storage.Write(finishedSession)
		if err != nil {
			s.log.Error(err)
		}
		delete(s.activeSessions, session.EditorID)
	}
}

// saveSession ends the current coding session and saves it to the filesystem.
func (s *Server) saveActiveSession() {
	if !s.activeSessions[s.activeEditorID].HasBuffers() {
		s.log.Debug("The session wasn't saved because it had not opened any buffers.",
			"editor_id", s.activeEditorID,
			"editor", s.activeSessions[s.activeEditorID].Editor,
			"os", s.activeSessions[s.activeEditorID].OS,
		)
		s.activeEditorID = ""
		delete(s.activeSessions, s.activeEditorID)
		return
	}

	s.log.Debug("Saving the session.",
		"editor_id", s.activeEditorID,
		"editor", s.activeSessions[s.activeEditorID].Editor,
		"os", s.activeSessions[s.activeEditorID].OS,
	)
	now := s.clock.GetTime()
	finishedSession := s.activeSessions[s.activeEditorID].End(now)
	err := s.storage.Write(finishedSession)
	if err != nil {
		s.log.Error(err)
	}

	delete(s.activeSessions, s.activeEditorID)
	s.activeEditorID = ""

	// Check if we should resume another session.
	if len(s.activeSessions) < 1 {
		return
	}

	var editorToResume string
	var mostRecentPause int64
	for _, session := range s.activeSessions {
		if session.PauseTime() > mostRecentPause {
			editorToResume = session.EditorID
			mostRecentPause = session.PauseTime()
		}
	}

	if editorToResume != "" {
		s.activeEditorID = editorToResume
		s.activeSessions[s.activeEditorID].Resume(s.clock.GetTime())
	}
}

func (s *Server) startServer(port string) (*http.Server, error) {
	proxy := proxy.New(s)
	err := rpc.RegisterName(s.name, proxy)
	if err != nil {
		return nil, err
	}

	rpc.HandleHTTP()
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return nil, err
	}
	httpServer := &http.Server{
		ReadHeaderTimeout: time.Second * 5,
	}
	go func() {
		serveErr := httpServer.Serve(listener)
		if !errors.Is(serveErr, http.ErrServerClosed) {
			panic(serveErr)
		}
	}()

	return httpServer, nil
}

// Start starts the server on the given port.
func (s *Server) Start(port string) error {
	s.log.Info("Starting up...")
	httpServer, err := s.startServer(port)
	if err != nil {
		return err
	}

	// Catch shutdown signals from the OS
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Blocks until a shutdown signal is received.
	sig := <-quit
	s.log.Info("Received shutdown signal", "signal", sig.String())

	// Stop the heartbeat checks and shutdown the http server.
	s.stopHeartbeatChecks <- struct{}{}
	err = httpServer.Shutdown(context.Background())
	if err != nil {
		s.log.Error(err)
		return err
	}

	// Save the all sessions before shutting down.
	s.saveAllSessions(s.clock.GetTime())
	s.log.Info("Shutting down.")

	return nil
}
