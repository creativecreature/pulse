package server_test

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"testing"
	"time"

	"github.com/charmbracelet/log"
	"github.com/creativecreature/pulse"
	"github.com/creativecreature/pulse/memory"
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

	mockStorage := memory.NewStorage()
	mockClock := pulse.NewTestClock(time.Now())

	reply := ""
	s, err := server.New("TestApp",
		server.WithLog(log.New(io.Discard)),
		server.WithStorage(mockStorage),
		server.WithClock(mockClock),
	)
	if err != nil {
		t.Fatal(err)
	}

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

	mockStorage := memory.NewStorage()
	mockClock := pulse.NewTestClock(time.Now())

	reply := ""
	s, err := server.New("TestApp",
		server.WithLog(log.New(io.Discard)),
		server.WithStorage(mockStorage),
		server.WithClock(mockClock),
	)
	if err != nil {
		t.Fatal(err)
	}

	// Open an initial VIM window.
	s.FocusGained(pulse.Event{
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
	mockClock.Add(100 * time.Millisecond)

	// Open the same file in a new editor instance.
	s.FocusGained(pulse.Event{
		EditorID: "345",
		Path:     "",
		Editor:   "nvim",
		OS:       "Linux",
	}, &reply)
	s.OpenFile(pulse.Event{
		EditorID: "345",
		Path:     absolutePath(t, "/testdata/sturdyc/cmd/main.go"),
		Editor:   "nvim",
		OS:       "Linux",
	}, &reply)
	// Push the clock forward to simulate that the file was opened for 50 ms.
	mockClock.Add(50 * time.Millisecond)

	// Open a third editor. This time, we'll never open a valid file.
	s.FocusGained(pulse.Event{
		EditorID: "678",
		Path:     "",
		Editor:   "nvim",
		OS:       "Linux",
	}, &reply)
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
	mockClock.Add(50 * time.Millisecond)

	// Open the first editor again, and close it.
	s.FocusGained(pulse.Event{
		EditorID: "123",
		Path:     absolutePath(t, "/testdata/sturdyc/cmd/main.go"),
		Editor:   "nvim",
		OS:       "Linux",
	}, &reply)
	mockClock.Add(10 * time.Millisecond)
	s.EndSession(pulse.Event{
		EditorID: "123",
		Path:     absolutePath(t, "/testdata/sturdyc/cmd/main.go"),
		Editor:   "nvim",
		OS:       "Linux",
	}, &reply)

	// This time should also be added towards the last active session.
	mockClock.Add(30 * time.Millisecond)

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
	mockClock.Add(500 * time.Millisecond)
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

func TestResumesThePreviousSession(t *testing.T) {
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

	mockStorage := memory.NewStorage()
	mockClock := pulse.NewTestClock(time.Now())

	reply := ""
	s, err := server.New("TestApp",
		server.WithLog(log.New(io.Discard)),
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
	mockClock.Add(100 * time.Millisecond)

	// Open the same file in a new editor instance.
	s.FocusGained(pulse.Event{
		EditorID: "345",
		Path:     "",
		Editor:   "nvim",
		OS:       "Linux",
	}, &reply)
	mockClock.Add(50 * time.Millisecond)
	s.OpenFile(pulse.Event{
		EditorID: "345",
		Path:     absolutePath(t, "/testdata/sturdyc/cmd/main.go"),
		Editor:   "nvim",
		OS:       "Linux",
	}, &reply)
	// Push the clock forward to simulate that the file was opened for 50 ms.
	mockClock.Add(50 * time.Millisecond)

	// Quit the session. It should resume the previous one.
	s.EndSession(pulse.Event{
		EditorID: "345",
		Path:     absolutePath(t, "/testdata/sturdyc/cmd/main.go"),
		Editor:   "nvim",
		OS:       "Linux",
	}, &reply)

	// This time should be added to the the last active session.
	mockClock.Add(30 * time.Millisecond)

	// Quit the first session.
	s.EndSession(pulse.Event{
		EditorID: "123",
		Path:     absolutePath(t, "/testdata/sturdyc/cmd/main.go"),
		Editor:   "nvim",
		OS:       "Linux",
	}, &reply)

	storedSessions, _ := mockStorage.Read()
	sort.Sort(storedSessions)
	if len(storedSessions) != 2 {
		t.Errorf("expected sessions %d; got %d", 2, len(storedSessions))
	}
	if storedSessions[0].DurationMs != 180 {
		t.Errorf("expected the sessions duration to be 180; got %d", storedSessions[0].DurationMs)
	}
	if storedSessions[1].DurationMs != 50 {
		t.Errorf("expected the sessions duration to be 50; got %d", storedSessions[1].DurationMs)
	}
}

func TestNoActivityShouldEndSession(t *testing.T) {
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

	mockStorage := memory.NewStorage()
	mockClock := pulse.NewTestClock(time.Now())

	reply := ""
	s, err := server.New("TestApp",
		server.WithLog(log.New(io.Discard)),
		server.WithStorage(mockStorage),
		server.WithClock(mockClock),
	)
	if err != nil {
		t.Fatal(err)
	}

	s.HeartbeatCheck()

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
	mockClock.Add(100 * time.Millisecond)

	// Open the same file in a new editor instance.
	s.OpenFile(pulse.Event{
		EditorID: "345",
		Path:     absolutePath(t, "/testdata/sturdyc/cmd/main.go"),
		Editor:   "nvim",
		OS:       "Linux",
	}, &reply)

	// Next, we'll wait for 10 minutes which should result in
	// both sessions being ended by a heartbeat check.
	mockClock.Add(10 * time.Minute)
	// Move the clock again to trigger the heartbeat ticker.
	mockClock.Add(time.Minute)
	time.Sleep(10 * time.Millisecond)

	storedSessions, _ := mockStorage.Read()
	err = mockStorage.Clean()
	if err != nil {
		t.Fatal(err)
	}
	sort.Sort(storedSessions)
	if len(storedSessions) != 2 {
		t.Errorf("expected sessions %d; got %d", 2, len(storedSessions))
	}

	if storedSessions[0].DurationMs != 100 {
		t.Errorf("expected the sessions duration to be 100; got %d", storedSessions[0].DurationMs)
	}

	// The second session should last for 11 minutes. 10 for the
	// heartbeat to expire, and 1 for the heartbeat to trigger.
	dur := int64(11 * time.Minute / time.Millisecond)
	if storedSessions[1].DurationMs != dur {
		t.Errorf("expected the sessions duration to be %d; got %d", dur, storedSessions[1].DurationMs)
	}
}
