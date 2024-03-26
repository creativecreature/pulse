package codeharvest_test

import (
	"testing"

	codeharvest "github.com/creativecreature/code-harvest"
)

func TestActiveSession(t *testing.T) {
	// Start a new coding session
	activeSession := codeharvest.StartSession("1337", 100, "linux", "nvim")

	// Open the first buffer
	bufferOne := codeharvest.NewBuffer(
		"init.lua",
		"dotfiles",
		"lua",
		"dotfiles/editors/nvim/init.lua",
		101,
	)
	activeSession.PushBuffer(bufferOne)

	// Open a second buffer.
	bufferTwo := codeharvest.NewBuffer(
		"plugins.lua",
		"dotfiles",
		"lua",
		"dotfiles/editors/nvim/plugins.lua",
		301,
	)
	activeSession.PushBuffer(bufferTwo)

	// Open the same file as buffer one. The total duration for these
	// buffers should be merged when we end the coding session.
	bufferThree := codeharvest.NewBuffer(
		"init.lua",
		"dotfiles",
		"lua",
		"dotfiles/editors/nvim/init.lua",
		611,
	)
	activeSession.PushBuffer(bufferThree)

	endedAt := int64(700)
	finishedSession := activeSession.End(endedAt)

	// Assert that the duration of the session was set correctly.
	if finishedSession.DurationMs != 600 {
		t.Errorf("Expected the session duration to be 600, got %d", finishedSession.DurationMs)
	}

	// Assert that the buffers have been merged into files.
	if len(finishedSession.Files) != 2 {
		t.Errorf("Expected the number of buffers to be 2, got %d", len(finishedSession.Files))
	}

	// Assert that the merged buffers has both durations.
	if finishedSession.Files[0].DurationMs != 289 {
		t.Errorf("Expected the merged duration for init.lua to be 289, got %d", finishedSession.Files[0].DurationMs)
	}

	if finishedSession.Files[1].DurationMs != 310 {
		t.Errorf("Expected the duration for plugins.lua to be 310, got %d", finishedSession.Files[1].DurationMs)
	}
}
