package handlers

import (
	"testing"
	"time"

	"poli-redi-api/internal/businessclock"
)

func TestAvailabilityRangeDefaultsAndLimits(t *testing.T) {
	if err := businessclock.Configure("America/Santiago"); err != nil {
		t.Fatal(err)
	}
	from, to, err := availabilityRange("", "")
	if err != nil {
		t.Fatal(err)
	}
	if got := int(to.Sub(from).Hours() / 24); got < 14 || got > 16 {
		t.Fatalf("default range days = %d", got)
	}
	if _, _, err := availabilityRange("2026-08-01", "2026-09-15"); err == nil {
		t.Fatal("range over 31 days was accepted")
	}
}

func TestAvailabilityRangeAcceptsRFC3339(t *testing.T) {
	from, to, err := availabilityRange("2026-08-14T10:00:00-04:00", "2026-08-14T12:00:00-04:00")
	if err != nil {
		t.Fatal(err)
	}
	if to.Sub(from) != 2*time.Hour {
		t.Fatalf("range duration = %s", to.Sub(from))
	}
}
