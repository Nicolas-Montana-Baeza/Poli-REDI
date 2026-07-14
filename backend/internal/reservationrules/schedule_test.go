package reservationrules

import (
	"strings"
	"testing"
	"time"
)

func TestValidateScheduleBoundaries(t *testing.T) {
	location, err := time.LoadLocation("America/Santiago")
	if err != nil {
		t.Fatalf("time.LoadLocation() error = %v", err)
	}

	tests := []struct {
		name     string
		hour     int
		minute   int
		second   int
		duration int
		wantErr  string
	}{
		{name: "opening", hour: 8, duration: 30},
		{name: "valid quarter hour", hour: 10, minute: 15, duration: 30},
		{name: "valid three-quarter hour", hour: 10, minute: 45, duration: 30},
		{name: "last valid block", hour: 21, minute: 30, duration: 30},
		{name: "three hours ending at close", hour: 19, duration: 180},
		{name: "before opening", hour: 7, minute: 30, duration: 30, wantErr: "08:00"},
		{name: "starts at close", hour: 22, duration: 30, wantErr: "anterior a las 22:00"},
		{name: "ends after close", hour: 21, minute: 30, duration: 60, wantErr: "finalizar"},
		{name: "invalid duration", hour: 10, duration: 45, wantErr: "duracion"},
		{name: "invalid minute step", hour: 10, minute: 10, duration: 30, wantErr: "intervalos"},
		{name: "invalid seconds", hour: 10, second: 30, duration: 30, wantErr: "intervalos"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			start := time.Date(2026, time.July, 20, test.hour, test.minute, test.second, 0, location)
			err := ValidateSchedule(start, test.duration)

			if test.wantErr == "" && err != nil {
				t.Fatalf("ValidateSchedule() error = %v", err)
			}

			if test.wantErr != "" && (err == nil || !strings.Contains(err.Error(), test.wantErr)) {
				t.Fatalf("ValidateSchedule() error = %v, expected %q", err, test.wantErr)
			}
		})
	}
}
