package pulse_test

import (
	"testing"
	"time"

	"github.com/creativecreature/pulse"
)

func TestMain(m *testing.M) {
	loc, err := time.LoadLocation("Europe/Stockholm")
	if err != nil {
		panic(err)
	}
	time.Local = loc

	m.Run()
}

func TestMergeSessions(t *testing.T) {
	t.Parallel()

	hourInMs := int64(60 * 60 * 1000)
	dayInMs := int64(24 * 60 * 60 * 1000)

	pulseFiles := pulse.Files{
		{Name: "main.go", Path: "cmd/main.go", Filetype: "go", DurationMs: 750},
		{Name: "logdb.go", Path: "logdb/logdb.go", Filetype: "go", DurationMs: 250},
	}
	pulseRepository := pulse.Repository{Name: "pulse", Files: pulseFiles, DurationMs: 1000}

	dotfilesFiles := pulse.Files{
		{Name: "init.lua", Path: "/nvim.init.lua", Filetype: "lua", DurationMs: 400},
		{Name: "install.sh", Path: "install.sh", Filetype: "shell", DurationMs: 500},
	}
	dotfilesRepository := pulse.Repository{Name: "dotfiles", Files: dotfilesFiles, DurationMs: 900}

	sturdycFiles := pulse.Files{
		{Name: "inflight.go", Path: "inflight.go", Filetype: "go", DurationMs: 2000},
		{Name: "keys.go", Path: "keys.go", Filetype: "go", DurationMs: 4000},
		{Name: "fetch.go", Path: "fetch.go", Filetype: "go", DurationMs: 100},
	}
	sturdycRepository := pulse.Repository{Name: "sturdyc", Files: sturdycFiles, DurationMs: 6100}

	sessions := pulse.CodingSessions{
		{
			ID:     "1",
			Period: pulse.Day,
			// 09:32 Friday June 16 2023
			EpochDateMs:  1686907956000,
			DateString:   "2023-06-16",
			TotalTimeMs:  1000,
			Repositories: pulse.Repositories{pulseRepository},
		},
		{
			ID:     "2",
			Period: pulse.Day,
			// 10:32 Friday June 16 2023
			EpochDateMs:  1686907956000 + hourInMs,
			DateString:   "2023-06-16",
			TotalTimeMs:  900,
			Repositories: pulse.Repositories{dotfilesRepository},
		},
		{
			ID:     "3",
			Period: pulse.Day,
			// 09:32 Saturday June 17 2023
			EpochDateMs:  1686907956000 + dayInMs,
			DateString:   "2023-06-17",
			TotalTimeMs:  900,
			Repositories: pulse.Repositories{sturdycRepository},
		},
		{
			ID:     "4",
			Period: pulse.Day,
			// 09:32 Tuesday June 21 2023
			EpochDateMs:  1686907956000 + dayInMs*4,
			DateString:   "2023-06-21",
			TotalTimeMs:  900,
			Repositories: pulse.Repositories{dotfilesRepository},
		},
		{
			ID:     "5",
			Period: pulse.Day,
			// 09:32 Friday July 28 2023
			EpochDateMs:  1686907956000 + dayInMs*42,
			DateString:   "2023-07-28",
			TotalTimeMs:  8000,
			Repositories: pulse.Repositories{pulseRepository, dotfilesRepository, sturdycRepository},
		},
		{
			ID:     "6",
			Period: pulse.Day,
			// 09:32 Saturday June 15 2024
			EpochDateMs:  1686907956000 + dayInMs*365,
			DateString:   "2024-06-15",
			TotalTimeMs:  8000,
			Repositories: pulse.Repositories{pulseRepository, dotfilesRepository, sturdycRepository},
		},
	}

	sessionsByDay := sessions.MergeByDay()
	sessionByWeek := sessions.MergeByWeek()
	sessionByMonth := sessions.MergeByMonth()
	sessionsByYear := sessions.MergeByYear()

	if len(sessionsByDay) != 5 {
		t.Errorf("expected 5 sessions, got %d", len(sessionsByDay))
	}

	if len(sessionByWeek) != 4 {
		t.Errorf("expected 4 sessions, got %d", len(sessionByWeek))
	}

	if len(sessionByMonth) != 3 {
		t.Errorf("expected 3 sessions, got %d", len(sessionByMonth))
	}

	if len(sessionsByYear) != 2 {
		t.Errorf("expected 2 sessions, got %d", len(sessionsByYear))
	}

	if sessionsByDay[0].TotalTimeMs != 1900 {
		t.Errorf("expected 1900 ms, got %d", sessionsByDay[0].TotalTimeMs)
	}

	if sessionsByDay[1].TotalTimeMs != 900 {
		t.Errorf("expected 900 ms, got %d", sessionsByDay[1].TotalTimeMs)
	}

	// 00:00 Friday June 16 2023
	if sessionsByDay[0].EpochDateMs != 1686873600000 {
		t.Errorf("expected 1686873600000, got %d", sessionsByDay[0].EpochDateMs)
	}

	if sessionByWeek[0].TotalTimeMs != 2800 {
		t.Errorf("expected 2800 ms, got %d", sessionByWeek[0].TotalTimeMs)
	}

	if sessionByWeek[1].TotalTimeMs != 900 {
		t.Errorf("expected 900 ms, got %d", sessionByWeek[1].TotalTimeMs)
	}

	// 00:00 Monday June 12 2023
	if sessionByWeek[0].EpochDateMs != 1686520800000 {
		t.Errorf("expected 1686520800000, got %d", sessionByWeek[0].EpochDateMs)
	}

	if sessionByMonth[0].TotalTimeMs != 3700 {
		t.Errorf("expected 3700 ms, got %d", sessionByMonth[0].TotalTimeMs)
	}

	if sessionByMonth[1].TotalTimeMs != 8000 {
		t.Errorf("expected 8000 ms, got %d", sessionByMonth[1].TotalTimeMs)
	}

	// 00:00 Thursday June 01 2023
	if sessionByMonth[0].EpochDateMs != 1685570400000 {
		t.Errorf("expected 1685570400000, got %d", sessionByMonth[0].EpochDateMs)
	}

	// 00:00 Saturday July 01 2024
	if sessionByMonth[1].EpochDateMs != 1688162400000 {
		t.Errorf("expected 1688162400000, got %d", sessionByMonth[1].EpochDateMs)
	}

	if sessionsByYear[0].TotalTimeMs != 11700 {
		t.Errorf("expected 11700 ms, got %d", sessionsByYear[0].TotalTimeMs)
	}

	if sessionsByYear[1].TotalTimeMs != 8000 {
		t.Errorf("expected 8000 ms, got %d", sessionsByYear[1].TotalTimeMs)
	}

	// 00:00 Sunday Jan 01 2023
	if sessionsByYear[0].EpochDateMs != 1672527600000 {
		t.Errorf("expected 1672527600000, got %d", sessionsByYear[0].EpochDateMs)
	}
}
