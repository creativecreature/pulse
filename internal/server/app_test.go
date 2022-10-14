package server_test

import (
	"io"
	"os"
	"testing"

	"code-harvest.conner.dev/internal/server"
	"code-harvest.conner.dev/internal/shared"
	"code-harvest.conner.dev/internal/storage"
	"code-harvest.conner.dev/pkg/logger"
)

func TestJumpingBetweenInstances(t *testing.T) {
	t.Parallel()

	log := logger.New(io.Discard, logger.LevelOff)
	storage := storage.MemoryStorage{}

	reply := ""
	s := server.New(log, &storage)

	// Open a new VIM instance
	s.FocusGained(shared.Event{
		Id:     "123",
		Path:   "",
		Editor: "nvim",
		OS:     "Linux",
	}, &reply)

	// Open a file in the first instance
	s.OpenFile(shared.Event{
		Id:     "123",
		Path:   "/Users/conner/code/creativecreature/dotfiles/install.sh",
		Editor: "nvim",
		OS:     "Linux",
	}, &reply)

	// Open another vim instance in a new split. This should end the previous session.
	s.FocusGained(shared.Event{
		Id:     "345",
		Path:   "",
		Editor: "nvim",
		OS:     "Linux",
	}, &reply)

	// Open a file in the second vim instance
	s.OpenFile(shared.Event{
		Id:     "345",
		Path:   "/Users/conner/code/creativecreature/dotfiles/bootstrap.sh",
		Editor: "nvim",
		OS:     "Linux",
	}, &reply)

	// Move focus back to the first VIM instance. This should end the second session.
	s.FocusGained(shared.Event{
		Id:     "123",
		Path:   "/Users/conner/code/creativecreature/dotfiles/install.sh",
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

	log := logger.New(os.Stdout, logger.LevelDebug)
	storage := storage.MemoryStorage{}

	reply := ""
	s := server.New(log, &storage)

	// Open a new instance of VIM
	s.FocusGained(shared.Event{
		Id:     "123",
		Path:   "",
		Editor: "nvim",
		OS:     "Linux",
	}, &reply)

	// Open a file
	s.OpenFile(shared.Event{
		Id:     "123",
		Path:   "/Users/conner/code/creativecreature/dotfiles/install.sh",
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

	// Lets now end the session. This behaviour should not have resulted in any
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
