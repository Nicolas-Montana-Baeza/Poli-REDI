package services

import (
	"strings"
	"testing"
	"time"

	"poli-redi-api/internal/businessclock"
	"poli-redi-api/internal/models"
	"poli-redi-api/internal/reservationrules"
)

func TestEnforceInitialReservationStatusIgnoresCallerStatus(t *testing.T) {
	for _, suppliedStatus := range []string{
		"",
		models.ReservationStatusPending,
		models.ReservationStatusCancelled,
		models.ReservationStatusRejected,
		models.ReservationStatusExpired,
	} {
		t.Run(suppliedStatus, func(t *testing.T) {
			reservation := enforceInitialReservationStatus(models.Reservation{
				Status: suppliedStatus,
			})

			if reservation.Status != models.ReservationStatusConfirmed {
				t.Fatalf("status = %q, expected %q", reservation.Status, models.ReservationStatusConfirmed)
			}
		})
	}
}

func TestValidateCancellationStatus(t *testing.T) {
	tests := []struct {
		status  string
		wantErr string
	}{
		{status: models.ReservationStatusConfirmed},
		{status: models.ReservationStatusPending},
		{status: models.ReservationStatusCancelled, wantErr: "ya est\u00e1 cancelada"},
		{status: models.ReservationStatusRejected, wantErr: "no se puede cancelar"},
		{status: models.ReservationStatusExpired, wantErr: "no se puede cancelar"},
	}

	for _, test := range tests {
		t.Run(test.status, func(t *testing.T) {
			err := validateCancellationStatus(test.status)

			if test.wantErr == "" && err != nil {
				t.Fatalf("validateCancellationStatus() error = %v", err)
			}

			if test.wantErr != "" && (err == nil || !strings.Contains(err.Error(), test.wantErr)) {
				t.Fatalf("validateCancellationStatus() error = %v, expected %q", err, test.wantErr)
			}
		})
	}
}

func TestCreateReservationAtRejectsPreviousBusinessDayNearMidnight(t *testing.T) {
	if err := businessclock.Configure(businessclock.DefaultLocationName); err != nil {
		t.Fatalf("Configure() error = %v", err)
	}

	now := time.Date(2026, time.July, 15, 0, 5, 0, 0, businessclock.Location())
	reservation := models.Reservation{
		UserID:          1,
		ResourceID:      1,
		StartTime:       time.Date(2026, time.July, 14, 23, 55, 0, 0, businessclock.Location()),
		DurationMinutes: 30,
	}

	_, err := createReservationAt(reservation, now)
	if err == nil || !strings.Contains(err.Error(), "pasado") {
		t.Fatalf("createReservationAt() error = %v, expected past reservation error", err)
	}
}

func TestDefaultScheduleRejectsInvalidValues(t *testing.T) {
	if err := businessclock.Configure(businessclock.DefaultLocationName); err != nil {
		t.Fatalf("Configure() error = %v", err)
	}

	tests := []struct {
		name     string
		start    time.Time
		duration int
		wantErr  string
	}{
		{
			name:     "before opening",
			start:    time.Date(2026, time.July, 20, 7, 30, 0, 0, businessclock.Location()),
			duration: 30,
			wantErr:  "08:00",
		},
		{
			name:     "manipulated duration",
			start:    time.Date(2026, time.July, 20, 10, 0, 0, 0, businessclock.Location()),
			duration: 45,
			wantErr:  "duracion",
		},
		{
			name:     "ends after closing",
			start:    time.Date(2026, time.July, 20, 21, 30, 0, 0, businessclock.Location()),
			duration: 60,
			wantErr:  "finalizar",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := reservationrules.ValidateSchedule(test.start, test.duration)

			if err == nil || !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("createReservationAt() error = %v, expected %q", err, test.wantErr)
			}
		})
	}
}

func TestValidateReservationPolicySnapshotUsesSuppliedVersion(t *testing.T) {
	location := time.FixedZone("America/Santiago", -4*60*60)
	now := time.Date(2026, 7, 20, 8, 0, 0, 0, location)
	reservation := models.Reservation{StartTime: time.Date(2026, 7, 24, 10, 0, 0, 0, location), DurationMinutes: 45}
	accepted := models.ReservationPolicy{ReservableWindowDays: 7, OpeningMinute: 8 * 60, ClosingMinute: 20 * 60, SlotIntervalMinutes: 15, AllowedDurations: []int{45}}
	if err := validateReservationPolicySnapshot(reservation, now, accepted); err != nil {
		t.Fatalf("supplied snapshot rejected reservation: %v", err)
	}
	rejected := accepted
	rejected.AllowedDurations = []int{30}
	if err := validateReservationPolicySnapshot(reservation, now, rejected); err == nil {
		t.Fatal("validation ignored the supplied policy snapshot")
	}
}

func TestRequestFrequencyMessageUsesSendDateEvenForLaterReservation(t *testing.T) {
	if err := businessclock.Configure(businessclock.DefaultLocationName); err != nil {
		t.Fatalf("Configure() error = %v", err)
	}
	previousCreatedAt := time.Date(2026, time.July, 20, 18, 30, 0, 0, time.UTC)
	now := time.Date(2026, time.July, 23, 10, 0, 0, 0, businessclock.Location())
	nextDate := reservationrules.NextRequestDate(previousCreatedAt, 7, businessclock.Location())
	selectedStart := time.Date(2026, time.July, 30, 12, 0, 0, 0, businessclock.Location())
	if !selectedStart.After(nextDate) {
		t.Fatal("test precondition: selected reservation must be later than next request date")
	}
	err := validateRequestFrequency(previousCreatedAt, 7, now)
	want := "A\u00fan no puedes enviar otra solicitud. Podr\u00e1s hacerlo desde el 27/07/2026; la espera depende de cu\u00e1ndo env\u00edas la solicitud, no de la fecha que quieres reservar."
	if err == nil || err.Error() != want {
		t.Fatalf("validateRequestFrequency() error = %q, expected %q", err, want)
	}
}

func TestOpenUseDoesNotConsumeRequestFrequency(t *testing.T) {
	if modeConsumesRequestFrequency("OPEN_USE") {
		t.Fatal("OPEN_USE must not consume request frequency")
	}
	for _, mode := range []string{"RESERVABLE", "ADMIN_ONLY"} {
		if !modeConsumesRequestFrequency(mode) {
			t.Fatalf("%s must consume request frequency", mode)
		}
	}
}
