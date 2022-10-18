package server_test

import (
	"io"
	"os"
	"testing"

	"code-harvest.conner.dev/internal/server"
	"code-harvest.conner.dev/internal/shared"
	"code-harvest.conner.dev/pkg/clock"
	"code-harvest.conner.dev/pkg/logger"
)

func TestJumpingBetweenInstances(t *testing.T) {
	t.Parallel()

	log := logger.New(os.Stdout, logger.LevelDebug)
	storage := server.MemoryStorage{}
	mockMetadataReader := &server.MockFileMetadataReader{}

	reply := ""
	s := server.New(log, &storage)
	s.MetadataReader = mockMetadataReader

	// Open a new VIM instance
	mockMetadataReader.Metadata = nil
	s.FocusGained(shared.Event{
		Id:     "123",
		Path:   "",
		Editor: "nvim",
		OS:     "Linux",
	}, &reply)

	// Open a file in the first instance
	mockMetadataReader.Metadata = &server.FileMetadata{
		Filename:       "install.sh",
		Filetype:       "bash",
		RepositoryName: "dotfiles",
	}
	s.OpenFile(shared.Event{
		Id:     "123",
		Path:   "/Users/conner/code/dotfiles/install.sh",
		Editor: "nvim",
		OS:     "Linux",
	}, &reply)

	// Open another vim instance in a new split. This should end the previous session.
	mockMetadataReader.Metadata = nil
	s.FocusGained(shared.Event{
		Id:     "345",
		Path:   "",
		Editor: "nvim",
		OS:     "Linux",
	}, &reply)

	// Open a file in the second vim instance
	mockMetadataReader.Metadata = &server.FileMetadata{
		Filename:       "bootstrap.sh",
		Filetype:       "bash",
		RepositoryName: "dotfiles",
	}
	s.OpenFile(shared.Event{
		Id:     "345",
		Path:   "/Users/conner/code/dotfiles/bootstrap.sh",
		Editor: "nvim",
		OS:     "Linux",
	}, &reply)

	// Move focus back to the first VIM instance. This should end the second session.
	mockMetadataReader.Metadata = &server.FileMetadata{
		Filename:       "install.sh",
		Filetype:       "bash",
		RepositoryName: "dotfiles",
	}
	s.FocusGained(shared.Event{
		Id:     "123",
		Path:   "/Users/conner/code/dotfiles/install.sh",
		Editor: "nvim",
		OS:     "Linux",
	}, &reply)

	// End the last session. We should now have 3 finished sessions.
	s.EndSession(shared.Event{
		Id:     "123",
		Path:   "",
		Editor: "nvim",
		OS:     "Linux",
	}, &reply)
	expectedNumberOfSessions := 3

	storedSessions := storage.Get()
	if len(storedSessions) != expectedNumberOfSessions {
		t.Errorf("expected len %d; got %d", expectedNumberOfSessions, len(storedSessions))
	}
}

func TestJumpBackAndForthToTheSameInstance(t *testing.T) {
	t.Parallel()

	log := logger.New(io.Discard, logger.LevelDebug)
	storage := server.MemoryStorage{}
	mockMetadataReader := &server.MockFileMetadataReader{}

	reply := ""
	s := server.New(log, &storage)
	s.MetadataReader = mockMetadataReader

	// Open a new instance of VIM
	mockMetadataReader.Metadata = nil
	s.FocusGained(shared.Event{
		Id:     "123",
		Path:   "",
		Editor: "nvim",
		OS:     "Linux",
	}, &reply)

	// Open a file
	mockMetadataReader.Metadata = &server.FileMetadata{
		Filename:       "install.sh",
		Filetype:       "bash",
		RepositoryName: "dotfiles",
	}
	s.OpenFile(shared.Event{
		Id:     "123",
		Path:   "/Users/conner/code/dotfiles/install.sh",
		Editor: "nvim",
		OS:     "Linux",
	}, &reply)

	// Lets now imagine we opened another TMUX split to run tests. We then jump
	// back to VIM which will fire the focus gained event with the same client id.
	s.FocusGained(shared.Event{
		Id:     "123",
		Path:   "",
		Editor: "nvim",
		OS:     "Linux",
	}, &reply)

	// We repeat the same thing again. Jump to another split in the terminal which makes
	// VIM lose focus and then back again - which will trigger another focus gained event.
	s.FocusGained(shared.Event{
		Id:     "123",
		Path:   "",
		Editor: "nvim",
		OS:     "Linux",
	}, &reply)
	mockMetadataReader.Metadata = &server.FileMetadata{
		Filename:       "bootstrap.sh",
		Filetype:       "bash",
		RepositoryName: "dotfiles",
	}
	s.OpenFile(shared.Event{
		Id:     "123",
		Path:   "/Users/conner/code/dotfiles/bootstrap.sh",
		Editor: "nvim",
		OS:     "Linux",
	}, &reply)

	// Lets now end the session. This behaviour should *not* have resulted in any
	// new sessions being created. We only create a new session and end the current
	// one if we open VIM in a new split (to not count double time).
	s.EndSession(shared.Event{
		Id:     "123",
		Path:   "",
		Editor: "nvim",
		OS:     "Linux",
	}, &reply)
	expectedNumberOfSessions := 1

	storedSessions := storage.Get()
	if len(storedSessions) != expectedNumberOfSessions {
		t.Errorf("expected len %d; got %d", expectedNumberOfSessions, len(storedSessions))
	}
}

func TestNoActivityShouldEndSession(t *testing.T) {
	t.Parallel()

	log := logger.New(os.Stdout, logger.LevelDebug)
	storage := server.MemoryStorage{}
	mockMetadataReader := &server.MockFileMetadataReader{}
	mockMetadataReader.Metadata = nil

	reply := ""
	s := server.New(log, &storage)

	mockClock := clock.MockClock{}
	s.Clock = &mockClock
	s.MetadataReader = mockMetadataReader

	// Send the initial focus event
	mockClock.SetTime(100)
	s.FocusGained(shared.Event{
		Id:     "123",
		Path:   "",
		Editor: "nvim",
		OS:     "Linux",
	}, &reply)

	mockClock.SetTime(200)
	s.CheckHeartbeat()

	// Send an open file event. This should update the time for the last activity to 250.
	mockClock.SetTime(250)
	mockMetadataReader.Metadata = &server.FileMetadata{
		Filename:       "install.sh",
		Filetype:       "bash",
		RepositoryName: "dotfiles",
	}
	s.OpenFile(shared.Event{
		Id:     "123",
		Path:   "/Users/conner/code/dotfiles/install.sh",
		Editor: "nvim",
		OS:     "Linux",
	}, &reply)

	// Perform another heartbeat check. Remember these checks does not update
	// the time for when we last saw activity in the session.
	mockClock.SetTime(300)
	s.CheckHeartbeat()

	// Heartbeat check that occurs 1 millisecond after the time of last activity
	// + ttl. This should result in the session being ended and saved.
	mockClock.SetTime(server.HeartbeatTTL.Milliseconds() + 250 + 1)
	s.CheckHeartbeat()

	mockClock.SetTime(server.HeartbeatTTL.Milliseconds() + 300)
	mockMetadataReader.Metadata = &server.FileMetadata{
		Filename:       "cleanup.sh",
		Filetype:       "bash",
		RepositoryName: "dotfiles",
	}
	s.OpenFile(shared.Event{
		Id:     "123",
		Path:   "/Users/conner/code/dotfiles/cleanup.sh",
		Editor: "nvim",
		OS:     "Linux",
	}, &reply)

	mockClock.SetTime(server.HeartbeatTTL.Milliseconds() + 400)
	s.CheckHeartbeat()

	s.EndSession(shared.Event{
		Id:     "123",
		Path:   "",
		Editor: "nvim",
		OS:     "Linux",
	}, &reply)

	expectedNumberOfSessions := 2
	storedSessions := storage.Get()
	if len(storedSessions) != expectedNumberOfSessions {
		t.Errorf("expected len %d; got %d", expectedNumberOfSessions, len(storedSessions))
	}
}
