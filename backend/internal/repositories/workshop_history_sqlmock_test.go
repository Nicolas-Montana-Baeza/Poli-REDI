package repositories

import (
	"regexp"
	"testing"
	"time"

	"poli-redi-api/internal/database"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestGetWorkshopEnrollmentsForUserIncludesInactiveAndCancelled(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	previous := database.DB
	database.DB = db
	defer func() { database.DB = previous }()

	enrolledAt := time.Date(2026, 7, 20, 15, 0, 0, 0, time.UTC)
	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT
			we.id,
			w.id,
			w.title,
			COALESCE(w.description, '') AS description,
			COALESCE(w.location, '') AS location,
			COALESCE(w.instructor_name, '') AS instructor_name,
			w.day_text,
			w.schedule_text,
			we.status,
			w.is_active,
			we.created_at
		FROM dbo.workshop_enrollments we
		INNER JOIN dbo.workshops w ON w.id = we.workshop_id
		WHERE we.user_id = @p1
		ORDER BY we.created_at DESC, we.id DESC;
		`)).
		WithArgs(42).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "workshop_id", "title", "description", "location", "instructor",
			"day", "schedule", "status", "active", "created_at",
		}).AddRow(
			8, 3, "Escalada", "", "Gimnasio", "Ana",
			"Martes", "18:00 - 19:00", "CANCELLED", false, enrolledAt,
		))

	items, err := GetWorkshopEnrollmentsForUser(42)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("got %d items, want 1", len(items))
	}
	if items[0].Status != "CANCELLED" || items[0].IsActive {
		t.Fatalf("inactive cancelled enrollment was not preserved: %#v", items[0])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
