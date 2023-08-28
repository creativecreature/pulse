package truncate_test

import (
	"testing"

	"github.com/creativecreature/code-harvest/truncate"
)

func TestTruncate(t *testing.T) {
	t.Parallel()

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

	actualDay := truncate.Day(originalTime)
	if actualDay != expectedDay {
		t.Errorf("Expected truncated day to be %d, got %d", expectedDay, actualDay)
	}

	actualWeek := truncate.Week(originalTime)
	if actualWeek != expectedWeek {
		t.Errorf("Expected truncated week to be %d, got %d", expectedWeek, actualWeek)
	}

	actualMonth := truncate.Month(originalTime)
	if actualMonth != expectedMonth {
		t.Errorf("Expected truncated month to be %d, got %d", expectedMonth, actualMonth)
	}

	actualYear := truncate.Year(originalTime)
	if actualYear != expectedYear {
		t.Errorf("Expected truncated year to be %d, got %d", expectedYear, actualYear)
	}
}
