package server

import (
	"context"
	"errors"
	"fmt"
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

const (
	HeartbeatTTL      = time.Minute * 10
	heartbeatInterval = time.Second * 45
)

type Server struct {
	name           string
	activeEditor   string
	activeSessions map[string]*pulse.CodingSession
	lastHeartbeat  int64
	clock          pulse.Clock
	fileReader     FileReader
	log            *log.Logger
	mutex          sync.Mutex
	storage        pulse.TemporaryStorage
}

// New creates a new server.
func New(serverName string, opts ...Option) (*Server, error) {
	a := &Server{
		name:           serverName,
		activeSessions: make(map[string]*pulse.CodingSession),
		clock:          pulse.NewClock(),
		fileReader:     filereader.New(),
	}
	for _, opt := range opts {
		err := opt(a)
		if err != nil {
			return &Server{}, err
		}
	}

	return a, nil
}

// startNewSession creates a new session and sets it as the current session.
func (s *Server) startNewSession(id, os, editor string) {
	s.log.Debug("Starting a new session")
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
	s.activeSessions[s.activeEditor].PushBuffer(buf)
	updatedBufferMsg := "Successfully updated the current buffer"
	s.log.Debug(updatedBufferMsg, "name", gitFile.Name, "path", gitFile.Path)
}

func (s *Server) saveAllSessions() {
	now := s.clock.GetTime()
	s.log.Debug("Saving all sessions.")

	for _, session := range s.activeSessions {
		if !session.HasBuffers() {
			s.log.Debug("The session has not opened any buffers.")
			return
		}

		finishedSession := session.End(now)
		err := s.storage.Write(finishedSession)
		if err != nil {
			s.log.Error(err)
		}
		delete(s.activeSessions, session.EditorID)
	}
}

// saveSession ends the current coding session and saves it to the filesystem.
func (s *Server) saveActiveSession() {
	if !s.activeSessions[s.activeEditor].HasBuffers() {
		s.log.Debug("The session wasn't saved because it had no open buffers.")
		return
	}

	s.log.Debug("Saving the session.")
	now := s.clock.GetTime()
	finishedSession := s.activeSessions[s.activeEditor].End(now)
	err := s.storage.Write(finishedSession)
	if err != nil {
		s.log.Error(err)
	}

	delete(s.activeSessions, s.activeEditor)
	s.activeEditor = ""

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
		s.activeEditor = editorToResume
		s.activeSessions[s.activeEditor].Resume(s.clock.GetTime())
	}
}

func (s *Server) startServer(port string) (*http.Server, error) {
	proxy := proxy.New(s)
	err := rpc.RegisterName(s.name, proxy)
	if err != nil {
		return nil, err
	}

	rpc.HandleHTTP()
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
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

// HeartbeatCheck runs a heartbeat ticker that ensures that
// the current session is not idle for more than ten minutes.
func (s *Server) HeartbeatCheck() func() {
	s.log.Info("Starting the ECG")
	ticker, stop := s.clock.NewTicker(heartbeatInterval)
	go func() {
		for range ticker {
			s.log.Debug("Checking the heartbeat")
			s.CheckHeartbeat()
		}
	}()

	return stop
}

// Start starts the server on the given port.
func (s *Server) Start(port string) error {
	s.log.Info("Starting up...")
	httpServer, err := s.startServer(port)
	if err != nil {
		return err
	}

	// Start the ECG. It will end inactive sessions.
	stopHeartbeat := s.HeartbeatCheck()

	// Catch shutdown signals from the OS
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Blocks until a shutdown signal is received.
	sig := <-quit
	s.log.Info("Received shutdown signal", "signal", sig.String())

	// Stop the heartbeat checks and shutdown the http server.
	stopHeartbeat()
	err = httpServer.Shutdown(context.Background())
	if err != nil {
		s.log.Error(err)
		return err
	}

	// Save the all sessions before shutting down.
	s.saveAllSessions()
	s.log.Info("Shutting down...")

	return nil
}
