package server_test

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/charmbracelet/log"
	"github.com/creativecreature/pulse"
	"github.com/creativecreature/pulse/clock"
	"github.com/creativecreature/pulse/server"
)

type mockStorage struct {
	sync.Mutex
	sessions []pulse.CodingSession
}

func newMockStorage() *mockStorage {
	return &mockStorage{
		sessions: make([]pulse.CodingSession, 0),
	}
}

func (m *mockStorage) Write(_ context.Context, session pulse.CodingSession) error {
	m.Lock()
	defer m.Unlock()
	m.sessions = append(m.sessions, session)
	return nil
}

func (m *mockStorage) GetSessions() []pulse.CodingSession {
	m.Lock()
	defer m.Unlock()
	return m.sessions
}

func absolutePath(t *testing.T, relativePath string) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("Could not get current file path")
	}
	return filepath.Join(filepath.Dir(filename), relativePath)
}

func TestServerMergesFiles(t *testing.T) {
	t.Parallel()

	// You can't commit a .git directory. Therefore, we have to rename it to .git in the test runner.
	err := os.Rename(absolutePath(t, "./testdata/sturdyc/git"), absolutePath(t, "./testdata/sturdyc/.git"))
	if err != nil {
		t.Fatal("Failed to set up .git directory for testing:", err)
	}
	defer func() {
		restoreErr := os.Rename(absolutePath(t, "./testdata/sturdyc/.git"), absolutePath(t, "./testdata/sturdyc/git"))
		if restoreErr != nil {
			t.Fatal("Failed to store the .git directory:", restoreErr)
		}
	}()

	mockClock := clock.NewMock(time.Now())
	mockStorage := newMockStorage()
	var cfg pulse.Config
	cfg.Server.Name = "TestApp"
	cfg.Server.AggregationInterval = 10 * time.Minute
	cfg.Server.SegmentationInterval = 5 * time.Minute
	cfg.Server.SegmentSizeKB = 10

	reply := ""
	s := server.New(&cfg, t.TempDir(), mockStorage,
		server.WithLog(log.New(io.Discard)),
		server.WithClock(mockClock),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		s.RunBackgroundJobs(ctx, cfg.Server.SegmentationInterval)
	}()
	time.Sleep(100 * time.Millisecond)

	// Open an initial VIM window.
	s.FocusGained(pulse.Event{
		EditorID: "123",
		Path:     "",
		Editor:   "nvim",
		OS:       "Linux",
	}, &reply)
	s.OpenFile(pulse.Event{
		EditorID: "123",
		Path:     "",
		Editor:   "nvim",
		OS:       "Linux",
	}, &reply)

	// Add some time between the session being started, and the first buffer opened.
	// Since this is the first session we started, the duration will still count
	// towards the total. It's only for new sessions that we require a valid
	// buffer to be opened for us to start counting time.
	mockClock.Add(10 * time.Millisecond)

	s.OpenFile(pulse.Event{
		EditorID: "123",
		Path:     absolutePath(t, "/testdata/sturdyc/cmd/main.go"),
		Editor:   "nvim",
		OS:       "Linux",
	}, &reply)
	// Push the clock forward to simulate that the file was opened for 100 ms.
	mockClock.Add(100 * time.Millisecond)

	// Open a second file.
	s.OpenFile(pulse.Event{
		EditorID: "123",
		Path:     absolutePath(t, "/testdata/sturdyc/pkg/foo/foo.go"),
		Editor:   "nvim",
		OS:       "Linux",
	}, &reply)
	// Push the clock forward to simulate that the file was opened for 50 ms.
	mockClock.Add(50 * time.Millisecond)

	// Open the first file again.
	s.OpenFile(pulse.Event{
		EditorID: "123",
		Path:     absolutePath(t, "/testdata/sturdyc/cmd/main.go"),
		Editor:   "nvim",
		OS:       "Linux",
	}, &reply)
	// Push the clock forward to simulate that the file was opened for 30 ms.
	mockClock.Add(30 * time.Millisecond)

	s.EndSession(pulse.Event{
		EditorID: "123",
		Path:     absolutePath(t, "/testdata/sturdyc/cmd/main.go"),
		Editor:   "nvim",
		OS:       "Linux",
	}, &reply)

	mockClock.Add(time.Minute * 30)
	mockClock.Add(time.Minute)
	time.Sleep(200 * time.Millisecond)

	storedSessions := mockStorage.GetSessions()
	if len(storedSessions) != 1 {
		t.Errorf("expected sessions %d; got %d", 1, len(storedSessions))
	}
	if storedSessions[0].Duration != 180*time.Millisecond {
		t.Errorf("expected the sessions duration to be 180 ms; got %d", storedSessions[0].Duration)
	}
	if storedSessions[0].Repositories[0].Duration != 180*time.Millisecond {
		t.Errorf("expected the repositories duration to be 180 ms; got %d", storedSessions[0].Repositories[0].Duration)
	}
	if len(storedSessions[0].Repositories[0].Files) != 2 {
		t.Errorf("expected the repositories files to be 2; got %d", len(storedSessions[0].Repositories[0].Files))
	}
}
