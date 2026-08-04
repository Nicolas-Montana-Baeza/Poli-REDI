package repositories

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestGetAvailabilityReservationsForRangeFiltersActiveOverlapsAndMarksParticipantConflict(t *testing.T) {
	mock, cleanup := withMockDB(t)
	defer cleanup()

	mock.ExpectQuery("SELECT r\\.id,r\\.user_id,r\\.start_time").
		WillReturnRows(sqlmock.NewRows([]string{"id", "owner", "start", "minutes", "minimum"}))
	columns := []string{
		"id", "policy_id", "user_id", "resource_id", "activity_id", "start_time",
		"duration", "status", "created_at", "updated_at", "activity", "resource",
		"full_name", "email", "rut", "target", "capacity", "minimum", "count",
		"deadline", "mode", "personal_conflict",
	}
	start := time.Date(2026, time.August, 4, 12, 0, 0, 0, time.UTC)
	mock.ExpectQuery("(?s)WHERE r\\.status IN \\('PENDING', 'CONFIRMED'\\).*r\\.start_time < @p2.*DATEADD\\(MINUTE, r\\.duration_minutes, r\\.start_time\\) > @p1").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), 7).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			1, 2, 99, 3, nil, start, 60, "CONFIRMED", start, start,
			"Basquetbol", "Cancha", "Owner", "owner@example.test", "1-9",
			10, 22, 10, 2, 60, "OPEN_USE", true,
		))

	items, err := GetAvailabilityReservationsForRange(start, start.Add(24*time.Hour), 7)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || !items[0].IsCurrentUserConflict || items[0].ReservationMode != "OPEN_USE" {
		t.Fatalf("unexpected items: %+v", items)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetAvailabilityBlocksForRangeFiltersActiveWithSemiOpenOverlap(t *testing.T) {
	mock, cleanup := withMockDB(t)
	defer cleanup()

	start := time.Date(2026, time.August, 4, 13, 0, 0, 0, time.UTC)
	mock.ExpectQuery("(?s)WHERE b\\.is_active = 1.*b\\.start_time < @p2 AND b\\.end_time > @p1").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "resource_id", "created_by", "type", "reason", "start", "end", "resource",
		}).AddRow(4, 2, 1, "MAINTENANCE", "detalle interno", start, start.Add(time.Hour), "Piscina"))

	items, err := GetAvailabilityBlocksForRange(start, start.Add(24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Reason != "detalle interno" {
		t.Fatalf("unexpected blocks: %+v", items)
	}
}

func TestGetWorkshopOccurrencesForRangeUsesNormalizedActiveSchedule(t *testing.T) {
	mock, cleanup := withMockDB(t)
	defer cleanup()

	start := time.Date(2026, time.August, 4, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery("(?s)WITH calendar_dates AS.*JOIN dbo\\.workshop_occurrences.*w\\.is_active = 1").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{
			"occurrence_id", "workshop_id", "resource_id", "title", "start", "end", "resource", "mode",
		}).AddRow(6, 5, 2, "Taller", start.Add(17*time.Hour), start.Add(18*time.Hour), "Cancha", "RESERVABLE"))

	items, err := GetWorkshopOccurrencesForRange(start, start.Add(48*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].WorkshopID != 5 || items[0].ReservationMode != "RESERVABLE" {
		t.Fatalf("unexpected workshops: %+v", items)
	}
}

func TestHasWorkshopAvailabilityConflictUsesSemiOpenComparison(t *testing.T) {
	mock, cleanup := withMockDB(t)
	defer cleanup()

	start := time.Date(2026, time.August, 4, 12, 0, 0, 0, time.UTC)
	mock.ExpectQuery("(?s)wo\\.start_minute < @p4.*@p3 < wo\\.end_minute").
		WithArgs(2, 2, 720, 780).
		WillReturnRows(sqlmock.NewRows([]string{"conflict"}).AddRow(false))

	conflict, err := HasWorkshopAvailabilityConflict(2, start, start.Add(time.Hour))
	if err != nil || conflict {
		t.Fatalf("conflict=%v err=%v", conflict, err)
	}
}
