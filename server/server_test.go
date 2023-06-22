package server_test

import (
	"io"
	"testing"

	"code-harvest.conner.dev/domain"
	"code-harvest.conner.dev/logger"
	"code-harvest.conner.dev/mock"
	"code-harvest.conner.dev/server"
	"code-harvest.conner.dev/storage"
)

func TestJumpingBetweenInstances(t *testing.T) {
	t.Parallel()

	mockStorage := storage.MemoryStorage()
	mockFileReader := mock.NewFileReader()

	a, err := server.New(
		"TestApp",
		server.WithLog(logger.New(io.Discard, logger.LevelOff)),
		server.WithFileReader(mockFileReader),
		server.WithStorage(mockStorage),
	)
	if err != nil {
		t.Fatal(err)
	}

	// Open a new VIM instance
	reply := ""
	a.FocusGained(domain.Event{
		Id:     "123",
		Path:   "",
		Editor: "nvim",
		OS:     "Linux",
	}, &reply)

	// Open a file in the first instance
	mockFileReader.SetFile(
		domain.GitFile{
			Name:       "install.sh",
			Filetype:   "bash",
			Repository: "dotfiles",
			Path:       "dotfiles/install.sh",
		},
	)
	a.OpenFile(domain.Event{
		Id:     "123",
		Path:   "/Users/conner/code/dotfiles/install.sh",
		Editor: "nvim",
		OS:     "Linux",
	}, &reply)

	// Open another vim instance in a new split. This should end the previous session.
	a.FocusGained(domain.Event{
		Id:     "345",
		Path:   "",
		Editor: "nvim",
		OS:     "Linux",
	}, &reply)

	// Open a file in the second vim instance
	mockFileReader.SetFile(
		domain.GitFile{
			Name:       "bootstrap.sh",
			Filetype:   "bash",
			Repository: "dotfiles",
			Path:       "dotfiles/bootstrap.sh",
		},
	)
	a.OpenFile(domain.Event{
		Id:     "345",
		Path:   "/Users/conner/code/dotfiles/bootstrap.sh",
		Editor: "nvim",
		OS:     "Linux",
	}, &reply)

	// Move focus back to the first VIM instance. This should end the second session.
	mockFileReader.SetFile(
		domain.GitFile{
			Name:       "install.sh",
			Filetype:   "bash",
			Repository: "dotfiles",
			Path:       "dotfiles/install.sh",
		},
	)
	a.FocusGained(domain.Event{
		Id:     "123",
		Path:   "/Users/conner/code/dotfiles/install.sh",
		Editor: "nvim",
		OS:     "Linux",
	}, &reply)

	// End the last session. We should now have 3 finished sessiona.
	a.EndSession(domain.Event{
		Id:     "123",
		Path:   "",
		Editor: "nvim",
		OS:     "Linux",
	}, &reply)

	expectedNumberOfSessions := 3
	storedSessions, _ := mockStorage.Read()

	if len(storedSessions) != expectedNumberOfSessions {
		t.Errorf("expected len %d; got %d", expectedNumberOfSessions, len(storedSessions))
	}
}

func TestJumpBackAndForthToTheSameInstance(t *testing.T) {
	t.Parallel()

	mockStorage := storage.MemoryStorage()
	mockFilereader := mock.NewFileReader()

	a, err := server.New(
		"testApp",
		server.WithLog(logger.New(io.Discard, logger.LevelOff)),
		server.WithFileReader(mockFilereader),
		server.WithStorage(mockStorage),
	)
	if err != nil {
		t.Fatal(err)
	}

	// Open a new instance of VIM
	reply := ""
	a.FocusGained(domain.Event{
		Id:     "123",
		Path:   "",
		Editor: "nvim",
		OS:     "Linux",
	}, &reply)

	// Open a file
	mockFilereader.SetFile(
		domain.GitFile{
			Name:       "install.sh",
			Filetype:   "bash",
			Repository: "dotfiles",
			Path:       "dotfiles/install.sh",
		},
	)
	a.OpenFile(domain.Event{
		Id:     "123",
		Path:   "/Users/conner/code/dotfiles/install.sh",
		Editor: "nvim",
		OS:     "Linux",
	}, &reply)

	// Lets now imagine we opened another TMUX split to run testa. We then jump
	// back to VIM which will fire the focus gained event with the same client id.
	a.FocusGained(domain.Event{
		Id:     "123",
		Path:   "",
		Editor: "nvim",
		OS:     "Linux",
	}, &reply)

	// We repeat the same thing again. Jump to another split in the terminal which makes
	// VIM lose focus and then back again - which will trigger another focus gained event.
	a.FocusGained(domain.Event{
		Id:     "123",
		Path:   "",
		Editor: "nvim",
		OS:     "Linux",
	}, &reply)
	mockFilereader.SetFile(
		domain.GitFile{
			Name:       "bootstrap.sh",
			Filetype:   "bash",
			Repository: "dotfiles",
			Path:       "dotfiles/bootstrap.sh",
		},
	)
	a.OpenFile(domain.Event{
		Id:     "123",
		Path:   "/Users/conner/code/dotfiles/bootstrap.sh",
		Editor: "nvim",
		OS:     "Linux",
	}, &reply)

	// Lets now end the session. This behaviour should *not* have resulted in any
	// new sessions being created. We only create a new session and end the current
	// one if we open VIM in a new split (to not count double time).
	a.EndSession(domain.Event{
		Id:     "123",
		Path:   "",
		Editor: "nvim",
		OS:     "Linux",
	}, &reply)

	expectedNumberOfSessions := 1
	storedSessions, _ := mockStorage.Read()

	if len(storedSessions) != expectedNumberOfSessions {
		t.Errorf("expected len %d; got %d", expectedNumberOfSessions, len(storedSessions))
	}
}

func TestNoActivityShouldEndSession(t *testing.T) {
	t.Parallel()

	mockStorage := storage.MemoryStorage()
	mockClock := &mock.Clock{}
	mockFilereader := mock.NewFileReader()

	a, err := server.New(
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
	a.FocusGained(domain.Event{
		Id:     "123",
		Path:   "",
		Editor: "nvim",
		OS:     "Linux",
	}, &reply)

	mockClock.SetTime(200)
	a.CheckHeartbeat()

	// Send an open file event. This should update the time for the last activity to 250.
	mockClock.SetTime(250)
	mockFilereader.SetFile(
		domain.GitFile{
			Name:       "install.sh",
			Filetype:   "bash",
			Repository: "dotfiles",
			Path:       "dotfiles/install.sh",
		},
	)
	a.OpenFile(domain.Event{
		Id:     "123",
		Path:   "/Users/conner/code/dotfiles/install.sh",
		Editor: "nvim",
		OS:     "Linux",
	}, &reply)

	// Perform another heartbeat check. Remember these checks does not update
	// the time for when we last saw activity in the session.
	mockClock.SetTime(300)
	a.CheckHeartbeat()

	// Heartbeat check that occurs 1 millisecond after the time of last activity
	// + ttl. This should result in the session being ended and saved.
	mockClock.SetTime(server.HeartbeatTTL.Milliseconds() + 250 + 1)
	a.CheckHeartbeat()

	mockClock.SetTime(server.HeartbeatTTL.Milliseconds() + 300)
	mockFilereader.SetFile(
		domain.GitFile{
			Name:       "cleanup.sh",
			Filetype:   "bash",
			Repository: "dotfiles",
			Path:       "dotfiles/cleanup.sh",
		},
	)
	a.OpenFile(domain.Event{
		Id:     "123",
		Path:   "/Users/conner/code/dotfiles/cleanup.sh",
		Editor: "nvim",
		OS:     "Linux",
	}, &reply)

	mockClock.SetTime(server.HeartbeatTTL.Milliseconds() + 400)
	a.CheckHeartbeat()

	a.EndSession(domain.Event{
		Id:     "123",
		Path:   "",
		Editor: "nvim",
		OS:     "Linux",
	}, &reply)

	expectedNumberOfSessions := 2
	storedSessions, _ := mockStorage.Read()

	if len(storedSessions) != expectedNumberOfSessions {
		t.Errorf("expected len %d; got %d", expectedNumberOfSessions, len(storedSessions))
	}
}
