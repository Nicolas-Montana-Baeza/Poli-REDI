package handlers

import (
	"database/sql"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gofiber/fiber/v2"
)

func workshopWithdrawalApp(user *models.LocalAuthUser) *fiber.App {
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		if user != nil {
			c.Locals("localUser", *user)
		}
		return c.Next()
	})
	app.Delete("/workshops/:id/enrollment", WithdrawFromWorkshop)
	return app
}

func withHandlerWorkshopDB(t *testing.T) (sqlmock.Sqlmock, func()) {
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

func expectHandlerWorkshopLocks(mock sqlmock.Sqlmock, workshopID, userID int, active bool) {
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id FROM dbo.users WITH \\(UPDLOCK, HOLDLOCK\\)").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(userID))
	mock.ExpectQuery("SELECT is_active[[:space:]]+FROM dbo.workshops WITH \\(UPDLOCK, HOLDLOCK\\)").
		WithArgs(workshopID).
		WillReturnRows(sqlmock.NewRows([]string{"is_active"}).AddRow(active))
}

func TestWorkshopWithdrawalRequiresAuthenticationAndValidID(t *testing.T) {
	response, err := workshopWithdrawalApp(nil).Test(
		httptest.NewRequest(http.MethodDelete, "/workshops/9/enrollment", nil),
	)
	if err != nil || response.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("unauthenticated status=%d err=%v", response.StatusCode, err)
	}

	response, err = workshopWithdrawalApp(&models.LocalAuthUser{ID: 42}).Test(
		httptest.NewRequest(http.MethodDelete, "/workshops/invalid/enrollment", nil),
	)
	if err != nil || response.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("invalid id status=%d err=%v", response.StatusCode, err)
	}
}

func TestWorkshopWithdrawalAllowsUserWithoutRUT(t *testing.T) {
	mock, cleanup := withHandlerWorkshopDB(t)
	defer cleanup()

	expectHandlerWorkshopLocks(mock, 9, 42, true)
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
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
	mock.ExpectCommit()

	app := workshopWithdrawalApp(&models.LocalAuthUser{ID: 42, RUT: ""})
	response, err := app.Test(httptest.NewRequest(http.MethodDelete, "/workshops/9/enrollment", nil))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != fiber.StatusOK || !strings.Contains(string(body), `"changed":true`) || !strings.Contains(string(body), `"enrolledCount":3`) {
		t.Fatalf("status=%d body=%s", response.StatusCode, body)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestWorkshopWithdrawalReturnsClosedAndNotFoundContracts(t *testing.T) {
	t.Run("closed", func(t *testing.T) {
		mock, cleanup := withHandlerWorkshopDB(t)
		defer cleanup()
		expectHandlerWorkshopLocks(mock, 9, 42, false)
		mock.ExpectRollback()

		response, err := workshopWithdrawalApp(&models.LocalAuthUser{ID: 42}).Test(
			httptest.NewRequest(http.MethodDelete, "/workshops/9/enrollment", nil),
		)
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		body, _ := io.ReadAll(response.Body)
		if response.StatusCode != fiber.StatusConflict || !strings.Contains(string(body), "WORKSHOP_ENROLLMENT_CLOSED") {
			t.Fatalf("status=%d body=%s", response.StatusCode, body)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("missing", func(t *testing.T) {
		mock, cleanup := withHandlerWorkshopDB(t)
		defer cleanup()
		mock.ExpectBegin()
		mock.ExpectQuery("SELECT id FROM dbo.users WITH \\(UPDLOCK, HOLDLOCK\\)").
			WithArgs(42).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(42))
		mock.ExpectQuery("SELECT is_active[[:space:]]+FROM dbo.workshops WITH \\(UPDLOCK, HOLDLOCK\\)").
			WithArgs(99).
			WillReturnError(sql.ErrNoRows)
		mock.ExpectRollback()

		response, err := workshopWithdrawalApp(&models.LocalAuthUser{ID: 42}).Test(
			httptest.NewRequest(http.MethodDelete, "/workshops/99/enrollment", nil),
		)
		if err != nil || response.StatusCode != fiber.StatusNotFound {
			t.Fatalf("status=%d err=%v", response.StatusCode, err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestWorkshopWithdrawalDoesNotLeakInternalErrors(t *testing.T) {
	mock, cleanup := withHandlerWorkshopDB(t)
	defer cleanup()
	mock.ExpectBegin().WillReturnError(errors.New("SECRET SQL DETAIL"))

	response, err := workshopWithdrawalApp(&models.LocalAuthUser{ID: 42}).Test(
		httptest.NewRequest(http.MethodDelete, "/workshops/9/enrollment", nil),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	body, _ := io.ReadAll(response.Body)
	if response.StatusCode != fiber.StatusInternalServerError || strings.Contains(string(body), "SECRET") {
		t.Fatalf("status=%d body=%s", response.StatusCode, body)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
