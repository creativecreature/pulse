package pulse_test

import (
	"testing"
	"time"

	"github.com/creativecreature/pulse"
)

func TestTruncate(t *testing.T) {
	t.Parallel()

	loc, err := time.LoadLocation("Europe/Stockholm")
	if err != nil {
		t.Fatal("Failed to load Stockholm timezone:", err)
	}
	time.Local = loc

	// 09:32 Friday June 16 2023
	originalTime := int64(1686907956000)
	// 00:00 Friday June 16 2023
	expectedDay := int64(1686873600000)
	// 00:00 Monday June 12 2023
	expectedWeek := int64(1686520800000)
	// 00:00 Thursday June 01 2023
	expectedMonth := int64(1685570400000)
	// 00:00 Sunday Jan 01 2023
	expectedYear := int64(1672527600000)

	actualDay := pulse.TruncateDay(originalTime)
	if actualDay != expectedDay {
		t.Errorf("Expected truncated day to be %d, got %d", expectedDay, actualDay)
	}

	actualWeek := pulse.TruncateWeek(originalTime)
	if actualWeek != expectedWeek {
		t.Errorf("Expected truncated week to be %d, got %d", expectedWeek, actualWeek)
	}

	actualMonth := pulse.TruncateMonth(originalTime)
	if actualMonth != expectedMonth {
		t.Errorf("Expected truncated month to be %d, got %d", expectedMonth, actualMonth)
	}

	actualYear := pulse.TruncateYear(originalTime)
	if actualYear != expectedYear {
		t.Errorf("Expected truncated year to be %d, got %d", expectedYear, actualYear)
	}
}
