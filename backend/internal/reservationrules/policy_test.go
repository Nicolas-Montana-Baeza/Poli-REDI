package reservationrules

import (
	"strings"
	"testing"
	"time"
)

func TestValidateReservableWindowUsesInclusiveInstitutionalDates(t *testing.T) {
	location, err := time.LoadLocation("America/Santiago")
	if err != nil {
		t.Fatal(err)
	}

	now := time.Date(2026, time.July, 21, 12, 0, 0, 0, location)
	for _, start := range []time.Time{
		time.Date(2026, time.July, 21, 8, 0, 0, 0, location),
		time.Date(2026, time.July, 27, 21, 30, 0, 0, location),
	} {
		if err := ValidateReservableWindow(now, start, 7); err != nil {
			t.Fatalf("ValidateReservableWindow() error = %v", err)
		}
	}

	err = ValidateReservableWindow(
		now,
		time.Date(2026, time.July, 28, 8, 0, 0, 0, location),
		7,
	)
	if err == nil || !strings.Contains(err.Error(), "2026-07-27") {
		t.Fatalf("ValidateReservableWindow() error = %v, expected inclusive limit", err)
	}
}

func TestValidateReservableWindowReflectsPolicyChange(t *testing.T) {
	location, _ := time.LoadLocation("America/Santiago")
	now := time.Date(2026, time.July, 21, 12, 0, 0, 0, location)
	start := time.Date(2026, time.July, 30, 10, 0, 0, 0, location)

	if err := ValidateReservableWindow(now, start, 7); err == nil {
		t.Fatal("expected seven-day policy to reject date")
	}
	if err := ValidateReservableWindow(now, start, 10); err != nil {
		t.Fatalf("ten-day policy error = %v", err)
	}
}

func TestNextRequestDateUsesCalendarDaysAcrossDST(t *testing.T) {
	location, _ := time.LoadLocation("America/Santiago")
	createdAt := time.Date(2026, time.September, 5, 15, 0, 0, 0, time.UTC)

	next := NextRequestDate(createdAt, 7, location)
	want := time.Date(2026, time.September, 12, 0, 0, 0, 0, location)

	if !next.Equal(want) {
		t.Fatalf("NextRequestDate() = %v, want %v", next, want)
	}
	if next.Sub(time.Date(2026, time.September, 5, 0, 0, 0, 0, location)) == 7*24*time.Hour {
		t.Fatal("expected calendar-day arithmetic to account for DST transition")
	}
}

func TestNextRequestDateAllowsSameWeekdayAfterConfiguredPeriod(t *testing.T) {
	location, _ := time.LoadLocation("America/Santiago")
	createdAt := time.Date(2026, time.July, 21, 15, 0, 0, 0, time.UTC)

	next := NextRequestDate(createdAt, 7, location)
	want := time.Date(2026, time.July, 28, 0, 0, 0, 0, location)

	if !next.Equal(want) {
		t.Fatalf("NextRequestDate() = %v, want %v", next, want)
	}
}
