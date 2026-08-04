package repositories

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"poli-redi-api/internal/database"

	"github.com/DATA-DOG/go-sqlmock"
)

func withWorkshopMockDB(t *testing.T) (sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	previous := database.DB
	database.DB = db
	return mock, func() {
		database.DB = previous
		db.Close()
	}
}

func expectWorkshopWithdrawalLocks(mock sqlmock.Sqlmock, workshopID, userID int, active bool) {
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id FROM dbo.users WITH \\(UPDLOCK, HOLDLOCK\\)").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(userID))
	mock.ExpectQuery("SELECT is_active[[:space:]]+FROM dbo.workshops WITH \\(UPDLOCK, HOLDLOCK\\)").
		WithArgs(workshopID).
		WillReturnRows(sqlmock.NewRows([]string{"is_active"}).AddRow(active))
}

func TestWithdrawUserFromWorkshopCancelsAndAuditsAtomically(t *testing.T) {
	mock, cleanup := withWorkshopMockDB(t)
	defer cleanup()

	expectWorkshopWithdrawalLocks(mock, 9, 42, true)
	mock.ExpectQuery("SELECT id[[:space:]]+FROM dbo.workshop_enrollments WITH \\(UPDLOCK, HOLDLOCK\\)").
		WithArgs(9, 42).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(55))
	mock.ExpectExec("UPDATE dbo.workshop_enrollments[[:space:]]+SET status = 'CANCELLED'").
		WithArgs(55).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO dbo.audit_logs.*WORKSHOP_ENROLLMENT_CANCELLED").
		WithArgs(42, 55).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery("SELECT COUNT\\(\\*\\)[[:space:]]+FROM dbo.workshop_enrollments WITH \\(UPDLOCK, HOLDLOCK\\)").
		WithArgs(9).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(4))
	mock.ExpectCommit()

	change, err := WithdrawUserFromWorkshop(9, 42)
	if err != nil {
		t.Fatal(err)
	}
	if !change.Changed || change.IsEnrolled || change.WorkshopID != 9 || change.EnrolledCount != 4 {
		t.Fatalf("unexpected change: %#v", change)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestWithdrawUserFromWorkshopIsIdempotent(t *testing.T) {
	mock, cleanup := withWorkshopMockDB(t)
	defer cleanup()

	expectWorkshopWithdrawalLocks(mock, 9, 42, true)
	mock.ExpectQuery("SELECT id[[:space:]]+FROM dbo.workshop_enrollments WITH \\(UPDLOCK, HOLDLOCK\\)").
		WithArgs(9, 42).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\)[[:space:]]+FROM dbo.workshop_enrollments WITH \\(UPDLOCK, HOLDLOCK\\)").
		WithArgs(9).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(4))
	mock.ExpectCommit()

	change, err := WithdrawUserFromWorkshop(9, 42)
	if err != nil {
		t.Fatal(err)
	}
	if change.Changed || change.IsEnrolled || change.EnrolledCount != 4 {
		t.Fatalf("unexpected idempotent change: %#v", change)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestWithdrawUserFromWorkshopRejectsInactiveAndMissingWorkshop(t *testing.T) {
	t.Run("inactive", func(t *testing.T) {
		mock, cleanup := withWorkshopMockDB(t)
		defer cleanup()
		expectWorkshopWithdrawalLocks(mock, 9, 42, false)
		mock.ExpectRollback()

		_, err := WithdrawUserFromWorkshop(9, 42)
		if !errors.Is(err, ErrWorkshopEnrollmentClosed) {
			t.Fatalf("got %v", err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("missing", func(t *testing.T) {
		mock, cleanup := withWorkshopMockDB(t)
		defer cleanup()
		mock.ExpectBegin()
		mock.ExpectQuery("SELECT id FROM dbo.users WITH \\(UPDLOCK, HOLDLOCK\\)").
			WithArgs(42).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(42))
		mock.ExpectQuery("SELECT is_active[[:space:]]+FROM dbo.workshops WITH \\(UPDLOCK, HOLDLOCK\\)").
			WithArgs(99).
			WillReturnError(sql.ErrNoRows)
		mock.ExpectRollback()

		_, err := WithdrawUserFromWorkshop(99, 42)
		if !errors.Is(err, ErrWorkshopNotFound) {
			t.Fatalf("got %v", err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestWithdrawUserFromWorkshopRollsBackWhenAuditFails(t *testing.T) {
	mock, cleanup := withWorkshopMockDB(t)
	defer cleanup()

	expectWorkshopWithdrawalLocks(mock, 9, 42, true)
	mock.ExpectQuery("SELECT id[[:space:]]+FROM dbo.workshop_enrollments WITH \\(UPDLOCK, HOLDLOCK\\)").
		WithArgs(9, 42).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(55))
	mock.ExpectExec("UPDATE dbo.workshop_enrollments[[:space:]]+SET status = 'CANCELLED'").
		WithArgs(55).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO dbo.audit_logs.*WORKSHOP_ENROLLMENT_CANCELLED").
		WithArgs(42, 55).
		WillReturnError(errors.New("audit unavailable"))
	mock.ExpectRollback()

	if _, err := WithdrawUserFromWorkshop(9, 42); err == nil {
		t.Fatal("expected audit failure")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestEnrollUserInWorkshopCreatesNewAuditedEpisode(t *testing.T) {
	mock, cleanup := withWorkshopMockDB(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id FROM dbo.users WITH \\(UPDLOCK, HOLDLOCK\\)").
		WithArgs(42).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(42))
	mock.ExpectQuery("SELECT[[:space:]]+w.capacity").
		WithArgs(9, 42).
		WillReturnRows(sqlmock.NewRows([]string{"capacity", "count", "enrolled"}).AddRow(20, 4, false))
	mock.ExpectQuery("SELECT COUNT\\(\\*\\)[[:space:]]+FROM dbo.workshop_occurrences").
		WithArgs(9).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT TOP \\(1\\) existing_w.title").
		WithArgs(42, 9).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery("DECLARE @created TABLE.*OUTPUT inserted.id INTO @created").
		WithArgs(9, 42).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(88))
	mock.ExpectExec("INSERT INTO dbo.audit_logs.*WORKSHOP_ENROLLMENT_CREATED").
		WithArgs(42, 88).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	now := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	mock.ExpectQuery("SELECT[[:space:]]+w.id").
		WithArgs(9, 42).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "resource", "title", "description", "location", "instructor", "day", "schedule",
			"capacity", "count", "active", "enrolled", "created", "updated",
		}).AddRow(9, 2, "Judo", "", "Gimnasio", "Ana", "Martes", "18:00 - 19:00", 20, 5, true, true, now, now))

	workshop, created, err := EnrollUserInWorkshop(9, 42)
	if err != nil {
		t.Fatal(err)
	}
	if !created || !workshop.IsEnrolled || workshop.EnrolledCount != 5 {
		t.Fatalf("unexpected enrollment result: created=%v workshop=%#v", created, workshop)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
