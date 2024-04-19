package pulse_test

import (
	"testing"

	"github.com/creativecreature/pulse/disk"
)

func TestReadAndAggregateSessions(t *testing.T) {
	t.Setenv("HOME", "testdata")
	storage, err := disk.NewStorage()
	if err != nil {
		t.Fatalf("Failed to create new storage: %v", err)
	}

	sessions, err := storage.Read()
	if err != nil {
		t.Fatalf("Failed to read sessions from disk: %v", err)
	}

	if len(sessions) != 8 {
		t.Errorf("Expected 8 sessions, got %d", len(sessions))
	}

	aggregatedSessions := sessions.Aggregate()
	if len(aggregatedSessions) != 3 {
		t.Errorf("Expected 3 aggregated sessions, got %d", len(aggregatedSessions))
	}

	dailySessions := aggregatedSessions.MergeByDay()
	if len(dailySessions) != 3 {
		t.Errorf("Expected 3 daily sessions, got %d", len(dailySessions))
	}

	monthlySessions := aggregatedSessions.MergeByMonth()
	if len(monthlySessions) != 2 {
		t.Errorf("Expected 2 monthly session, got %d", len(monthlySessions))
	}

	yearlySessions := aggregatedSessions.MergeByYear()
	if len(yearlySessions) != 1 {
		t.Errorf("Expected 1 yearly session, got %d", len(yearlySessions))
	}
}
