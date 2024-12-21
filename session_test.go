package pulse_test

import (
	"testing"
	"time"

	"github.com/viccon/pulse"
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
		{Name: "main.go", Path: "cmd/main.go", Filetype: "go", Duration: 750},
		{Name: "logdb.go", Path: "logdb/logdb.go", Filetype: "go", Duration: 250},
	}
	pulseRepository := pulse.Repository{Name: "pulse", Files: pulseFiles, Duration: 1000}

	dotfilesFiles := pulse.Files{
		{Name: "init.lua", Path: "/nvim.init.lua", Filetype: "lua", Duration: 400},
		{Name: "install.sh", Path: "install.sh", Filetype: "shell", Duration: 500},
	}
	dotfilesRepository := pulse.Repository{Name: "dotfiles", Files: dotfilesFiles, Duration: 900}

	sturdycFiles := pulse.Files{
		{Name: "inflight.go", Path: "inflight.go", Filetype: "go", Duration: 2000},
		{Name: "keys.go", Path: "keys.go", Filetype: "go", Duration: 4000},
		{Name: "fetch.go", Path: "fetch.go", Filetype: "go", Duration: 100},
	}
	sturdycRepository := pulse.Repository{Name: "sturdyc", Files: sturdycFiles, Duration: 6100}

	sessions := pulse.CodingSessions{
		{
			// 09:32 Friday June 16 2023
			Date:         time.UnixMilli(1686907956000),
			Duration:     1000,
			Repositories: pulse.Repositories{pulseRepository},
		},
		{
			// 10:32 Friday June 16 2023
			Date:         time.UnixMilli(1686907956000 + hourInMs),
			Duration:     900,
			Repositories: pulse.Repositories{dotfilesRepository},
		},
		{
			// 09:32 Saturday June 17 2023
			Date:         time.UnixMilli(1686907956000 + dayInMs),
			Duration:     900,
			Repositories: pulse.Repositories{sturdycRepository},
		},
		{
			// 09:32 Tuesday June 21 2023
			Date:         time.UnixMilli(1686907956000 + dayInMs*4),
			Duration:     900,
			Repositories: pulse.Repositories{dotfilesRepository},
		},
		{
			// 09:32 Friday July 28 2023
			Date:         time.UnixMilli(1686907956000 + dayInMs*42),
			Duration:     8000,
			Repositories: pulse.Repositories{pulseRepository, dotfilesRepository, sturdycRepository},
		},
		{
			// 09:32 Saturday June 15 2024
			Date:         time.UnixMilli(1686907956000 + dayInMs*365),
			Duration:     8000,
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

	if sessionsByDay[0].Duration != 1900 {
		t.Errorf("expected 1900, got %d", sessionsByDay[0].Duration)
	}

	if sessionsByDay[1].Duration != 900 {
		t.Errorf("expected 900, got %d", sessionsByDay[1].Duration)
	}

	// 00:00 Friday June 16 2023
	if sessionsByDay[0].Date.UnixMilli() != 1686866400000 {
		t.Errorf("expected 1686866400000, got %d", sessionsByDay[0].Date.UnixMilli())
	}

	if sessionByWeek[0].Duration != 2800 {
		t.Errorf("expected 2800, got %d", sessionByWeek[0].Duration)
	}

	if sessionByWeek[1].Duration != 900 {
		t.Errorf("expected 900, got %d", sessionByWeek[1].Duration)
	}

	// 00:00 Monday June 12 2023
	if sessionByWeek[0].Date.UnixMilli() != 1686520800000 {
		t.Errorf("expected 1686520800000, got %d", sessionByWeek[0].Date.UnixMilli())
	}

	if sessionByMonth[0].Duration != 3700 {
		t.Errorf("expected 3700, got %d", sessionByMonth[0].Duration)
	}

	if sessionByMonth[1].Duration != 8000 {
		t.Errorf("expected 8000, got %d", sessionByMonth[1].Duration)
	}

	// 00:00 Thursday June 01 2023
	if sessionByMonth[0].Date.UnixMilli() != 1685570400000 {
		t.Errorf("expected 1685570400000, got %d", sessionByMonth[0].Date.UnixMilli())
	}

	// 00:00 Saturday July 01 2024
	if sessionByMonth[1].Date.UnixMilli() != 1688162400000 {
		t.Errorf("expected 1688162400000, got %d", sessionByMonth[1].Date.UnixMilli())
	}

	if sessionsByYear[0].Duration != 11700 {
		t.Errorf("expected 11700, got %d", sessionsByYear[0].Duration)
	}

	if sessionsByYear[1].Duration != 8000 {
		t.Errorf("expected 8000, got %d", sessionsByYear[1].Duration)
	}

	// 00:00 Sunday Jan 01 2023
	if sessionsByYear[0].Date.UnixMilli() != 1672527600000 {
		t.Errorf("expected 1672527600000, got %d", sessionsByYear[0].Date.UnixMilli())
	}
}
