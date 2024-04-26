package pulse_test

import (
	"testing"
	"time"

	"github.com/creativecreature/pulse"
)

func TestActiveSession(t *testing.T) {
	t.Parallel()

	mockClock := pulse.NewTestClock(time.Now())

	// Start a new coding session
	activeSession := pulse.StartSession("1337", mockClock.Now(), "linux", "nvim")

	// Open the first buffer, and wait 400ms.
	bufferOne := pulse.NewBuffer(
		"init.lua",
		"dotfiles",
		"lua",
		"dotfiles/editors/nvim/init.lua",
		mockClock.Now(),
	)
	activeSession.PushBuffer(bufferOne)
	mockClock.Add(400 * time.Millisecond)

	// Open a second buffer, and wait 200ms.
	bufferTwo := pulse.NewBuffer(
		"plugins.lua",
		"dotfiles",
		"lua",
		"dotfiles/editors/nvim/plugins.lua",
		mockClock.Now(),
	)
	activeSession.PushBuffer(bufferTwo)
	mockClock.Add(200 * time.Millisecond)

	// Open the first buffer again. The total duration for these
	// buffers should be merged when we end the coding session.
	bufferThree := pulse.NewBuffer(
		"init.lua",
		"dotfiles",
		"lua",
		"dotfiles/editors/nvim/init.lua",
		mockClock.Now(),
	)
	activeSession.PushBuffer(bufferThree)
	mockClock.Add(time.Millisecond * 100)

	finishedSession := activeSession.End(mockClock.Now())

	// Assert that the duration of the session was set correctly.
	if finishedSession.Duration.Milliseconds() != 700 {
		t.Errorf("Expected the session duration to be 600, got %d", finishedSession.Duration.Milliseconds())
	}

	// Assert that the buffers have been merged into files.
	if len(finishedSession.Files) != 2 {
		t.Errorf("Expected the number of buffers to be 2, got %d", len(finishedSession.Files))
	}

	// Assert that the merged buffers has both durations.
	initLuaDuration := finishedSession.Files[0].Duration.Milliseconds()
	if initLuaDuration != 500 {
		t.Errorf("Expected the merged duration for init.lua to be 500, got %d", initLuaDuration)
	}

	pluginsLuaDuration := finishedSession.Files[1].Duration.Milliseconds()
	if pluginsLuaDuration != 200 {
		t.Errorf("Expected the duration for plugins.lua to be 200, got %d", pluginsLuaDuration)
	}
}
