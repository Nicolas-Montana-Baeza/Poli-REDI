package services

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"poli-redi-api/internal/models"
)

func intPointer(value int) *int {
	return &value
}

func timePointer(value time.Time) *time.Time {
	return &value
}

func TestReservationAvailabilityKindDoesNotDependOnStatus(t *testing.T) {
	groupTarget := 10

	tests := []struct {
		name        string
		reservation models.Reservation
		want        string
	}{
		{
			name:        "confirmed ordinary reservation",
			reservation: models.Reservation{Status: models.ReservationStatusConfirmed},
			want:        models.AvailabilityKindReservation,
		},
		{
			name:        "pending group reservation",
			reservation: models.Reservation{Status: models.ReservationStatusPending, TargetParticipants: &groupTarget},
			want:        models.AvailabilityKindGroupReservation,
		},
		{
			name:        "confirmed group reservation",
			reservation: models.Reservation{Status: models.ReservationStatusConfirmed, TargetParticipants: &groupTarget},
			want:        models.AvailabilityKindGroupReservation,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := reservationAvailabilityKind(test.reservation); got != test.want {
				t.Fatalf("reservationAvailabilityKind() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestSanitizeAvailabilityItemsKeepsOwnReservationWithoutPII(t *testing.T) {
	deadline := time.Date(2099, time.July, 30, 18, 0, 0, 0, time.UTC)
	input := []models.AvailabilityItem{{
		ID:                   1,
		AvailabilityKind:     models.AvailabilityKindGroupReservation,
		UserID:               7,
		Title:                "Basquetbol",
		UserFullName:         "Usuario siete",
		UserEmail:            "siete@example.test",
		UserRUT:              "11111111-1",
		ParticipantCount:     3,
		MinimumParticipants:  10,
		TargetParticipants:   intPointer(12),
		Capacity:             intPointer(22),
		ConfirmationDeadline: timePointer(deadline),
		CanEditTarget:        true,
	}}

	got := SanitizeAvailabilityItemsForAudience(input, models.LocalAuthUser{ID: 7})

	if got[0].UserID != 7 || got[0].Title != "Basquetbol" {
		t.Fatalf("own reservation lost its safe detail: %+v", got[0])
	}
	if got[0].ParticipantCount != 3 || got[0].TargetParticipants == nil || !got[0].CanEditTarget {
		t.Fatalf("own reservation lost its management detail: %+v", got[0])
	}
	if got[0].UserFullName != "" || got[0].UserEmail != "" || got[0].UserRUT != "" {
		t.Fatalf("availability returned unnecessary PII for the owner: %+v", got[0])
	}
}

func TestSanitizeAvailabilityItemsRedactsForeignGroupButKeepsSafeKind(t *testing.T) {
	deadline := time.Date(2099, time.July, 30, 18, 0, 0, 0, time.UTC)
	input := []models.AvailabilityItem{{
		ID:                   2,
		AvailabilityKind:     models.AvailabilityKindGroupReservation,
		UserID:               99,
		Title:                "Actividad privada",
		UserFullName:         "Propietario secreto",
		UserEmail:            "owner@example.test",
		UserRUT:              "22222222-2",
		ParticipantCount:     8,
		MinimumParticipants:  10,
		TargetParticipants:   intPointer(14),
		Capacity:             intPointer(22),
		ConfirmationDeadline: timePointer(deadline),
		CanEditTarget:        true,
	}}

	got := SanitizeAvailabilityItemsForAudience(input, models.LocalAuthUser{ID: 7})
	item := got[0]

	if item.AvailabilityKind != models.AvailabilityKindGroupReservation {
		t.Fatalf("safe group kind was lost: %+v", item)
	}
	if item.Title != "Reserva" || item.UserID != 0 ||
		item.UserFullName != "" || item.UserEmail != "" || item.UserRUT != "" {
		t.Fatalf("foreign identity or activity leaked: %+v", item)
	}
	if item.ParticipantCount != 0 || item.MinimumParticipants != 0 ||
		item.TargetParticipants != nil || item.Capacity != nil ||
		item.ConfirmationDeadline != nil || item.CanEditTarget {
		t.Fatalf("foreign group metrics leaked: %+v", item)
	}

	payload, err := json.Marshal(item)
	if err != nil {
		t.Fatal(err)
	}
	body := string(payload)
	for _, forbidden := range []string{
		"userId", "userFullName", "userEmail", "userRut",
		"Actividad privada", "Propietario secreto", "owner@example.test", "22222222-2",
		"participantCount", "minimumParticipants", "targetParticipants",
		"capacity", "confirmationDeadline",
	} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("sanitized JSON still contains %q: %s", forbidden, body)
		}
	}
}

func TestSanitizeAvailabilityItemsKeepsScheduledCategoryButNotInternalDetail(t *testing.T) {
	input := []models.AvailabilityItem{{
		ID:                  3,
		AvailabilityKind:    models.AvailabilityKindScheduled,
		UserID:              42,
		Title:               "Entrenamiento selección universitaria",
		UserFullName:        "Administrador",
		UserEmail:           "admin@example.test",
		UserRUT:             "33333333-3",
		IsScheduledActivity: true,
		ActivityType:        "TRAINING",
	}}

	got := SanitizeAvailabilityItemsForAudience(input, models.LocalAuthUser{ID: 7})
	item := got[0]

	if item.Title != "Actividad institucional" ||
		item.ActivityType != "TRAINING" ||
		item.AvailabilityKind != models.AvailabilityKindScheduled {
		t.Fatalf("scheduled activity was not sanitized by contract: %+v", item)
	}
	if item.UserID != 0 || item.UserFullName != "" || item.UserEmail != "" || item.UserRUT != "" {
		t.Fatalf("scheduled activity creator leaked: %+v", item)
	}
}

func TestSanitizeAvailabilityItemsLeavesAdministratorPayloadUnchanged(t *testing.T) {
	input := []models.AvailabilityItem{{
		ID:                   4,
		AvailabilityKind:     models.AvailabilityKindGroupReservation,
		UserID:               99,
		Title:                "Basquetbol",
		UserFullName:         "Usuario",
		UserEmail:            "user@example.test",
		UserRUT:              "44444444-4",
		ParticipantCount:     5,
		MinimumParticipants:  10,
		TargetParticipants:   intPointer(12),
		Capacity:             intPointer(22),
		ConfirmationDeadline: timePointer(time.Date(2099, time.July, 30, 18, 0, 0, 0, time.UTC)),
		CanEditTarget:        true,
	}}

	got := SanitizeAvailabilityItemsForAudience(input, models.LocalAuthUser{ID: 1, IsAdmin: true})

	if !reflect.DeepEqual(got, input) {
		t.Fatalf("administrator payload changed:\n got: %+v\nwant: %+v", got, input)
	}
}

func TestSanitizeAvailabilityItemsDoesNotMutateSourceSlice(t *testing.T) {
	input := []models.AvailabilityItem{{
		ID:               5,
		AvailabilityKind: models.AvailabilityKindReservation,
		UserID:           99,
		Title:            "Actividad original",
		UserEmail:        "owner@example.test",
	}}

	_ = SanitizeAvailabilityItemsForAudience(input, models.LocalAuthUser{ID: 7})

	if input[0].Title != "Actividad original" || input[0].UserEmail != "owner@example.test" {
		t.Fatalf("source slice was mutated: %+v", input[0])
	}
}
