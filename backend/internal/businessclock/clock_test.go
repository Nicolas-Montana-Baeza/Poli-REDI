package businessclock

import (
	"testing"
	"time"
)

func TestParseDateTimeUsesBusinessTimezoneForWallTime(t *testing.T) {
	if err := Configure(DefaultLocationName); err != nil {
		t.Fatalf("Configure() error = %v", err)
	}

	parsed, err := ParseDateTime("2026-07-14T10:30:00")
	if err != nil {
		t.Fatalf("ParseDateTime() error = %v", err)
	}

	_, offset := parsed.Zone()
	if parsed.Hour() != 10 || offset != -4*60*60 {
		t.Fatalf("ParseDateTime() = %v, expected 10:30 in winter Santiago time", parsed)
	}
}

func TestParseDateTimeHonorsExplicitOffset(t *testing.T) {
	if err := Configure(DefaultLocationName); err != nil {
		t.Fatalf("Configure() error = %v", err)
	}

	parsed, err := ParseDateTime("2026-07-14T14:30:00Z")
	if err != nil {
		t.Fatalf("ParseDateTime() error = %v", err)
	}

	if parsed.Hour() != 10 || parsed.Minute() != 30 {
		t.Fatalf("ParseDateTime() = %v, expected 10:30 in Santiago", parsed)
	}
}

func TestDatabaseWallTimeRoundTripPreservesInstitutionalHour(t *testing.T) {
	if err := Configure(DefaultLocationName); err != nil {
		t.Fatalf("Configure() error = %v", err)
	}

	local := time.Date(2026, time.December, 10, 18, 45, 0, 0, Location())
	databaseValue := ToDatabaseWallTime(local)
	roundTrip := FromDatabaseWallTime(databaseValue)

	if !roundTrip.Equal(local) {
		t.Fatalf("round trip = %v, expected %v", roundTrip, local)
	}

	if databaseValue.Hour() != 18 || databaseValue.Location() != time.UTC {
		t.Fatalf("database value = %v, expected UTC-tagged wall time 18:45", databaseValue)
	}
}

func TestConfigureRejectsInvalidTimezone(t *testing.T) {
	if err := Configure("Mars/Olympus"); err == nil {
		t.Fatal("Configure() expected an error for an invalid timezone")
	}

	if err := Configure(DefaultLocationName); err != nil {
		t.Fatalf("restore Configure() error = %v", err)
	}
}
