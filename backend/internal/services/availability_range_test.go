package services

import (
	"testing"
	"time"

	"poli-redi-api/internal/models"
)

func TestBuildAvailabilityItemsAppliesSourceAndResourceModeMatrix(t *testing.T) {
	start := time.Date(2026, time.August, 4, 12, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	items := buildAvailabilityItems(
		[]models.AvailabilityReservation{
			{Reservation: models.Reservation{ID: 1, ResourceID: 1, StartTime: start, DurationMinutes: 60, Status: models.ReservationStatusConfirmed, Type: "normal"}, ReservationMode: "RESERVABLE", IsCurrentUserConflict: true},
			{Reservation: models.Reservation{ID: 2, ResourceID: 2, StartTime: start, DurationMinutes: 60, Status: models.ReservationStatusConfirmed, Type: "normal"}, ReservationMode: "OPEN_USE"},
		},
		[]models.ScheduledActivity{
			{ID: 3, ResourceID: 3, StartTime: start, EndTime: end, ReservationMode: "RESERVABLE"},
			{ID: 4, ResourceID: 4, StartTime: start, EndTime: end, ReservationMode: "OPEN_USE"},
		},
		[]models.WorkshopAvailabilityOccurrence{
			{ID: 5, ResourceID: 5, StartTime: start, EndTime: end, ReservationMode: "RESERVABLE"},
			{ID: 6, ResourceID: 6, StartTime: start, EndTime: end, ReservationMode: "OPEN_USE"},
		},
		[]models.AvailabilityBlock{
			{ID: 7, ResourceID: 7, StartTime: start, EndTime: end},
		},
	)

	byResource := make(map[int]models.AvailabilityItem, len(items))
	for _, item := range items {
		byResource[item.ResourceID] = item
	}

	for _, resourceID := range []int{1, 3, 5, 7} {
		if !byResource[resourceID].BlocksResource {
			t.Fatalf("resource %d should be blocked: %+v", resourceID, byResource[resourceID])
		}
	}
	for _, resourceID := range []int{2, 4, 6} {
		if byResource[resourceID].BlocksResource {
			t.Fatalf("OPEN_USE resource %d should not be blocked: %+v", resourceID, byResource[resourceID])
		}
	}
	if !byResource[1].IsCurrentUserConflict {
		t.Fatal("owner/confirmed participant reservation must be a personal conflict")
	}
	for _, resourceID := range []int{3, 4, 5, 6, 7} {
		if byResource[resourceID].IsCurrentUserConflict {
			t.Fatalf("non-reservation source %d must not create a personal conflict", resourceID)
		}
	}
	if byResource[1].AvailabilityKind != models.AvailabilityKindReservation ||
		byResource[3].AvailabilityKind != models.AvailabilityKindScheduled ||
		byResource[5].AvailabilityKind != models.AvailabilityKindWorkshop ||
		byResource[7].AvailabilityKind != models.AvailabilityKindBlock {
		t.Fatalf("unexpected kinds: %+v", byResource)
	}
}

func TestSanitizeAvailabilityItemsProtectsBlockAndForeignReservationButKeepsWorkshopTitle(t *testing.T) {
	items := []models.AvailabilityItem{
		{AvailabilityKind: models.AvailabilityKindBlock, UserID: 90, Title: "Fuga en camarines", ActivityType: "MAINTENANCE"},
		{AvailabilityKind: models.AvailabilityKindWorkshop, UserID: 91, Title: "Taller de esgrima"},
		{AvailabilityKind: models.AvailabilityKindReservation, UserID: 92, Title: "Actividad privada", IsCurrentUserConflict: true},
	}

	got := SanitizeAvailabilityItemsForAudience(items, models.LocalAuthUser{ID: 7})
	if got[0].Title != "No disponible" || got[0].ActivityType != "" || got[0].UserID != 0 {
		t.Fatalf("block leaked sensitive detail: %+v", got[0])
	}
	if got[1].Title != "Taller de esgrima" || got[1].UserID != 0 {
		t.Fatalf("workshop public title was not preserved safely: %+v", got[1])
	}
	if got[2].Title != "Reserva" || !got[2].IsCurrentUserConflict {
		t.Fatalf("foreign reservation privacy/conflict contract changed: %+v", got[2])
	}
}
