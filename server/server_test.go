package server_test

import (
	"io"
	"path/filepath"
	"runtime"
	"sort"
	"testing"

	"github.com/creativecreature/pulse"
	"github.com/creativecreature/pulse/logger"
	"github.com/creativecreature/pulse/memory"
	"github.com/creativecreature/pulse/mock"
	"github.com/creativecreature/pulse/server"
)

func absolutePath(t *testing.T, relativePath string) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("Could not get current file path")
	}
	return filepath.Join(filepath.Dir(filename), relativePath)
}

func TestServerMergesFiles(t *testing.T) {
	mockStorage := memory.NewStorage()
	mockClock := &mock.Clock{}
	mockClock.SetTime(0)

	reply := ""
	s, err := server.New("TestApp",
		server.WithLog(logger.New(io.Discard, logger.LevelOff)),
		server.WithStorage(mockStorage),
		server.WithClock(mockClock),
	)
	if err != nil {
		t.Fatal(err)
	}

	// Open an initial VIM window.
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
	mockClock.AddTime(10)

	s.OpenFile(pulse.Event{
		EditorID: "123",
		Path:     absolutePath(t, "/testdata/sturdyc/cmd/main.go"),
		Editor:   "nvim",
		OS:       "Linux",
	}, &reply)
	// Push the clock forward to simulate that the file was opened for 100 ms.
	mockClock.AddTime(100)

	// Open a second file.
	s.OpenFile(pulse.Event{
		EditorID: "123",
		Path:     absolutePath(t, "/testdata/sturdyc/pkg/foo/foo.go"),
		Editor:   "nvim",
		OS:       "Linux",
	}, &reply)
	// Push the clock forward to simulate that the file was opened for 50 ms.
	mockClock.AddTime(50)

	// Open the first file again.
	s.OpenFile(pulse.Event{
		EditorID: "123",
		Path:     absolutePath(t, "/testdata/sturdyc/cmd/main.go"),
		Editor:   "nvim",
		OS:       "Linux",
	}, &reply)
	// Push the clock forward to simulate that the file was opened for 30 ms.
	mockClock.AddTime(30)

	s.EndSession(pulse.Event{
		EditorID: "123",
		Path:     absolutePath(t, "/testdata/sturdyc/cmd/main.go"),
		Editor:   "nvim",
		OS:       "Linux",
	}, &reply)

	storedSessions, _ := mockStorage.Read()
	sort.Sort(storedSessions)
	if len(storedSessions) != 1 {
		t.Errorf("expected sessions %d; got %d", 1, len(storedSessions))
	}
	if storedSessions[0].DurationMs != 190 {
		t.Errorf("expected the sessions duration to be 190; got %d", storedSessions[0].DurationMs)
	}
	if len(storedSessions[0].Files) != 2 {
		t.Errorf("expected files %d; got %d", 2, len(storedSessions[0].Files))
	}
	if storedSessions[0].Files[0].DurationMs != 130 {
		t.Errorf("expected file duration 130; got %d", storedSessions[0].Files[0].DurationMs)
	}
	if storedSessions[0].Files[1].DurationMs != 50 {
		t.Errorf("expected file duration 50; got %d", storedSessions[0].Files[1].DurationMs)
	}
	if storedSessions[0].Files[0].Path != "sturdyc/cmd/main.go" {
		t.Errorf("expected file path sturdyc/cmd/main.go; got %s", storedSessions[0].Files[0].Path)
	}
	if storedSessions[0].Files[1].Path != "sturdyc/pkg/foo/foo.go" {
		t.Errorf("expected file path sturdyc/pkg/foo/foo.go; got %s", storedSessions[0].Files[1].Path)
	}
	if storedSessions[0].Files[0].Repository != "sturdyc" {
		t.Errorf("expected file repository sturdyc; got %s", storedSessions[0].Files[0].Repository)
	}
	if storedSessions[0].Files[1].Repository != "sturdyc" {
		t.Errorf("expected file repository sturdyc; got %s", storedSessions[0].Files[1].Repository)
	}
}

func TestTimeGetsAddedToTheCorrectSession(t *testing.T) {
	mockStorage := memory.NewStorage()
	mockClock := &mock.Clock{}
	mockClock.SetTime(0)

	reply := ""
	s, err := server.New("TestApp",
		server.WithLog(logger.New(io.Discard, logger.LevelOff)),
		server.WithStorage(mockStorage),
		server.WithClock(mockClock),
	)
	if err != nil {
		t.Fatal(err)
	}

	// Open an initial VIM window.
	s.OpenFile(pulse.Event{
		EditorID: "123",
		Path:     "",
		Editor:   "nvim",
		OS:       "Linux",
	}, &reply)

	s.OpenFile(pulse.Event{
		EditorID: "123",
		Path:     absolutePath(t, "/testdata/sturdyc/cmd/main.go"),
		Editor:   "nvim",
		OS:       "Linux",
	}, &reply)
	// Push the clock forward to simulate that the file was opened for 100 ms.
	mockClock.AddTime(100)

	// Open the same file in a new editor instance.
	s.OpenFile(pulse.Event{
		EditorID: "345",
		Path:     absolutePath(t, "/testdata/sturdyc/cmd/main.go"),
		Editor:   "nvim",
		OS:       "Linux",
	}, &reply)
	// Push the clock forward to simulate that the file was opened for 50 ms.
	mockClock.AddTime(50)

	// Open a third editor. This time, we'll never open a valid file.
	s.OpenFile(pulse.Event{
		EditorID: "678",
		Path:     "",
		Editor:   "nvim",
		OS:       "Linux",
	}, &reply)
	s.OpenFile(pulse.Event{
		EditorID: "678",
		// This is a temporary buffer without a file.
		Path:   absolutePath(t, "/testdata/sturdyc/cmd/NvimTree_1"),
		Editor: "nvim",
		OS:     "Linux",
	}, &reply)
	// Given that we haven't opened a valid file yet, the time should count to the previous session.
	mockClock.AddTime(50)

	// Open the first editor again, and close it.
	s.OpenFile(pulse.Event{
		EditorID: "123",
		Path:     absolutePath(t, "/testdata/sturdyc/cmd/main.go"),
		Editor:   "nvim",
		OS:       "Linux",
	}, &reply)
	mockClock.AddTime(10)
	s.EndSession(pulse.Event{
		EditorID: "123",
		Path:     absolutePath(t, "/testdata/sturdyc/cmd/main.go"),
		Editor:   "nvim",
		OS:       "Linux",
	}, &reply)

	// This time should also be added towards the last active session.
	mockClock.AddTime(30)

	// Open the editor with ID 345 again, and close it.
	s.OpenFile(pulse.Event{
		EditorID: "345",
		Path:     absolutePath(t, "/testdata/sturdyc/cmd/main.go"),
		Editor:   "nvim",
		OS:       "Linux",
	}, &reply)
	s.EndSession(pulse.Event{
		EditorID: "345",
		Path:     absolutePath(t, "/testdata/sturdyc/cmd/main.go"),
		Editor:   "nvim",
		OS:       "Linux",
	}, &reply)

	// Open editor 678 again. Remember, this editor has not opened any
	// valid files. Adding time, and then closing it should be a no-op.
	s.OpenFile(pulse.Event{
		EditorID: "678",
		Path:     "",
		Editor:   "nvim",
		OS:       "Linux",
	}, &reply)
	mockClock.AddTime(500)
	s.EndSession(pulse.Event{
		EditorID: "678",
		Path:     absolutePath(t, "/testdata/sturdyc/cmd/main.go"),
		Editor:   "nvim",
		OS:       "Linux",
	}, &reply)

	storedSessions, _ := mockStorage.Read()
	sort.Sort(storedSessions)
	if len(storedSessions) != 2 {
		t.Errorf("expected sessions %d; got %d", 2, len(storedSessions))
	}
	if storedSessions[0].DurationMs != 110 {
		t.Errorf("expected the sessions duration to be 110; got %d", storedSessions[0].DurationMs)
	}
	if storedSessions[1].DurationMs != 130 {
		t.Errorf("expected the sessions duration to be 130; got %d", storedSessions[1].DurationMs)
	}
}

func TestNoActivityShouldEndSession(t *testing.T) {
	mockStorage := memory.NewStorage()
	mockClock := &mock.Clock{}
	mockFilereader := mock.NewFileReader()

	s, err := server.New(
		"testApp",
		server.WithLog(logger.New(io.Discard, logger.LevelOff)),
		server.WithClock(mockClock),
		server.WithFileReader(mockFilereader),
		server.WithStorage(mockStorage),
	)
	if err != nil {
		t.Fatal(err)
	}

	// Send the initial focus event
	mockClock.SetTime(100)
	reply := ""
	s.FocusGained(pulse.Event{
		EditorID: "123",
		Path:     "",
		Editor:   "nvim",
		OS:       "Linux",
	}, &reply)
	if err != nil {
		panic(err)
	}

	mockClock.SetTime(200)
	s.CheckHeartbeat()

	// Send an open file event. This should update the time for the last activity to 250.
	mockClock.SetTime(250)
	mockFilereader.SetFile(
		pulse.GitFile{
			Name:       "install.sh",
			Filetype:   "bash",
			Repository: "dotfiles",
			Path:       "dotfiles/install.sh",
		},
	)

	s.OpenFile(pulse.Event{
		EditorID: "123",
		Path:     "/Users/conner/code/dotfiles/install.sh",
		Editor:   "nvim",
		OS:       "Linux",
	}, &reply)

	// Perform another heartbeat check.
	mockClock.SetTime(300)
	s.CheckHeartbeat()

	//  Move the clock again passed the heartbeat TTL.
	mockClock.SetTime(server.HeartbeatTTL.Milliseconds() + 250 + 1)
	s.CheckHeartbeat()

	mockClock.SetTime(server.HeartbeatTTL.Milliseconds() + 300)
	mockFilereader.SetFile(
		pulse.GitFile{
			Name:       "cleanup.sh",
			Filetype:   "bash",
			Repository: "dotfiles",
			Path:       "dotfiles/cleanup.sh",
		},
	)
	s.OpenFile(pulse.Event{
		EditorID: "123",
		Path:     "/Users/conner/code/dotfiles/cleanup.sh",
		Editor:   "nvim",
		OS:       "Linux",
	}, &reply)
	mockClock.SetTime(server.HeartbeatTTL.Milliseconds() + 400)
	s.CheckHeartbeat()

	s.EndSession(pulse.Event{
		EditorID: "123",
		Path:     "",
		Editor:   "nvim",
		OS:       "Linux",
	}, &reply)

	storedSessions, _ := mockStorage.Read()

	expectedNumberOfSessions := 2
	if len(storedSessions) != expectedNumberOfSessions {
		t.Errorf("expected len %d; got %d", expectedNumberOfSessions, len(storedSessions))
	}
}
