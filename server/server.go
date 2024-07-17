package server

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/rpc"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/creativecreature/pulse"
	"github.com/creativecreature/pulse/clock"
	"github.com/creativecreature/pulse/git"
	"github.com/creativecreature/pulse/logdb"
	"github.com/creativecreature/pulse/logger"
)

type SessionWriter interface {
	Write(context.Context, pulse.CodingSession) error
}

type Server struct {
	cfg           *pulse.Config
	clock         clock.Clock
	logDB         *logdb.LogDB
	logger        *log.Logger
	mu            sync.Mutex
	activeBuffer  *pulse.Buffer
	lastHeartbeat time.Time
	sessionWriter SessionWriter
}

// New creates a new server.
func New(cfg *pulse.Config, segmentPath string, sessionWriter SessionWriter, opts ...Option) *Server {
	s := &Server{
		cfg:           cfg,
		clock:         clock.New(),
		logger:        logger.New(),
		sessionWriter: sessionWriter,
	}

	for _, opt := range opts {
		opt(s)
	}

	s.logDB = logdb.NewDB(segmentPath, cfg.Server.SegmentSizeKB, s.clock)

	return s
}

func (s *Server) openFile(event pulse.Event) {
	gitFile, gitFileErr := git.ParseFile(event.Path)
	if gitFileErr != nil {
		return
	}

	if s.activeBuffer != nil {
		if s.activeBuffer.Filepath == gitFile.Path && s.activeBuffer.Repository == gitFile.Repository {
			s.logger.Debug("This buffer is already considered active",
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

	s.logger.Debug("Writing the buffer")
	buf := s.activeBuffer
	buf.Close(s.clock.Now())
	key := buf.Key()

	// Merge the duration with the most recent entry for this day.
	if bytes, hasMostRecentEntry := s.logDB.Get(key); hasMostRecentEntry {
		s.logger.Debug("Merging with the most recent entry for this buffer")
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
	s.logDB.MustSet(key, bytes)
	s.activeBuffer = nil
}

// RunBackgroundJobs starts the heartbeat, aggregation, and segmentation jobs.
func (s *Server) RunBackgroundJobs(ctx context.Context, segmentationInterval time.Duration) {
	go s.runHeartbeatChecks(ctx)
	go s.runAggregations(ctx)
	go s.logDB.RunSegmentations(ctx, segmentationInterval)
}

// Start starts the server on the given port.
func (s *Server) StartServer(ctx context.Context, port string) error {
	s.logger.Info("Starting up...")
	proxy := NewProxy(s)
	err := rpc.RegisterName(s.cfg.Server.Name, proxy)
	if err != nil {
		return err
	}

	rpc.HandleHTTP()
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
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

	// Blocks until the context is cancelled.
	<-ctx.Done()
	s.logger.Info("Shutting down")
	s.saveBuffer()

	shutdownContext, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	//nolint: contextcheck // This is a new cancellation tree.
	return httpServer.Shutdown(shutdownContext)
}
