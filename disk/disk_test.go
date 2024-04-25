package disk_test

import (
	"os"
	"path"
	"testing"
	"time"

	"github.com/creativecreature/pulse"
	"github.com/creativecreature/pulse/disk"

	"github.com/google/go-cmp/cmp"
)

func TestNewStorageCreatesPulseDir(t *testing.T) {
	// Create a temporary directory to simulate the home directory
	tempHome := t.TempDir()

	// Set the HOME environment variable to the temporary directory
	t.Setenv("HOME", tempHome)

	storage, err := disk.NewStorage()
	if err != nil {
		t.Fatalf("Failed to create new storage: %v", err)
	}

	// Check if the .pulse directory was created correctly
	expectedPath := path.Join(tempHome, ".pulse")
	if _, statErr := os.Stat(expectedPath); os.IsNotExist(statErr) {
		t.Errorf(".pulse directory was not created in the home directory")
	}

	if storage.Root() != expectedPath {
		t.Errorf("Storage root expected %v, got %v", expectedPath, storage.Root())
	}
}

func TestStorageReadWriteClean(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	storage, err := disk.NewStorage()
	if err != nil {
		t.Fatalf("Failed to create new storage: %v", err)
	}

	now := time.Now()
	sessions := pulse.Sessions{
		pulse.Session{
			StartedAt: now,
			EndedAt:   now.Add(time.Hour),
			Duration:  time.Hour,
			OS:        "linux",
			Editor:    "nvim",
			Files: pulse.Files{
				pulse.File{
					Name:       "main.go",
					Path:       "/cmd/main.go",
					Repository: "pulse",
					Filetype:   "go",
					Duration:   time.Hour,
				},
			},
		},
		pulse.Session{
			StartedAt: now.Add(time.Minute),
			EndedAt:   now.Add(time.Minute * 11),
			Duration:  time.Hour,
			OS:        "linux",
			Editor:    "nvim",
			Files: pulse.Files{
				pulse.File{
					Name:       "main.go",
					Path:       "/cmd/main.go",
					Repository: "pulse",
					Filetype:   "go",
					Duration:   time.Minute * 10,
				},
			},
		},
	}

	for _, session := range sessions {
		writeErr := storage.Write(session)
		if writeErr != nil {
			t.Fatalf("Failed to write session to disk: %v", writeErr)
		}
	}

	storedSessions, readErr := storage.Read()
	if readErr != nil {
		t.Fatalf("Failed to read sessions from disk: %v", readErr)
	}

	if !cmp.Equal(sessions, storedSessions) {
		t.Error(cmp.Diff(sessions, storedSessions))
	}

	cleanErr := storage.Clean()
	if cleanErr != nil {
		t.Fatalf("Failed to clean sessions from disk: %v", cleanErr)
	}

	storedSessions, readErr = storage.Read()
	if readErr != nil {
		t.Fatalf("Failed to read sessions from disk: %v", readErr)
	}
	if len(storedSessions) != 0 {
		t.Errorf("Clean did not remove all sessions from disk")
	}
}
