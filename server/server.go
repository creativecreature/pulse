package server

import (
	"context"
	"encoding/json"
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
	"github.com/creativecreature/pulse/git"
	"github.com/creativecreature/pulse/logdb"
)

type CodingSessionWriter interface {
	Write(context.Context, pulse.CodingSession) error
}

type Server struct {
	name          string
	activeBuffer  *pulse.Buffer
	lastHeartbeat time.Time
	stopJobs      chan struct{}
	clock         Clock
	log           *log.Logger
	mutex         sync.Mutex
	db            *logdb.LogDB
	sessionWriter CodingSessionWriter
}

// New creates a new server.
func New(serverName, segmentPath string, sessionWriter CodingSessionWriter, opts ...Option) (*Server, error) {
	s := &Server{
		name:          serverName,
		clock:         NewClock(),
		stopJobs:      make(chan struct{}),
		db:            logdb.New(segmentPath),
		sessionWriter: sessionWriter,
	}
	for _, opt := range opts {
		err := opt(s)
		if err != nil {
			return &Server{}, err
		}
	}

	// Run the heartbeat checks and aggregations in the background.
	s.runHeartbeatChecks()
	s.runAggregations()

	return s, nil
}

func (s *Server) openFile(event pulse.Event) {
	gitFile, gitFileErr := git.ParseFile(event.Path)
	if gitFileErr != nil {
		return
	}

	if s.activeBuffer != nil {
		if s.activeBuffer.Filepath == gitFile.Path && s.activeBuffer.Repository == gitFile.Repository {
			s.log.Debug("This buffer is already considered active.",
				"path", gitFile.Path,
				"repository", gitFile.Repository,
				"editor_id", event.EditorID,
				"editor", event.Editor,
				"os", event.OS,
			)
			return
		}
	}

	s.saveBuffer()
	buf := pulse.NewBuffer(
		gitFile.Name,
		gitFile.Repository,
		gitFile.Filetype,
		gitFile.Path,
		s.clock.Now(),
	)
	s.activeBuffer = &buf
}

// saveBuffer writes the currently open buffer to disk. Should be called with a lock.
func (s *Server) saveBuffer() {
	if s.activeBuffer == nil {
		return
	}

	s.log.Debug("Writing the buffer.")
	buf := s.activeBuffer
	buf.Close(s.clock.Now())
	key := buf.Key()

	// Merge the duration with the most recent entry for this day.
	if bytes, hasMostRecentEntry := s.db.Get(key); hasMostRecentEntry {
		var b pulse.Buffer
		err := json.Unmarshal(bytes, &b)
		if err != nil {
			panic(err)
		}
		buf.Duration += b.Duration
	}

	bytes, err := json.Marshal(buf)
	if err != nil {
		panic(err)
	}
	s.db.MustSet(key, bytes)
	s.activeBuffer = nil
}

func (s *Server) startServer(port string) (*http.Server, error) {
	proxy := NewProxy(s)
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
	s.stopJobs <- struct{}{}
	err = httpServer.Shutdown(context.Background())
	if err != nil {
		s.log.Error(err)
		return err
	}

	s.saveBuffer()
	s.log.Info("Shutting down.")

	return nil
}
