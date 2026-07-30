package handlers

import (
	"database/sql"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"
	"poli-redi-api/internal/repositories"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gofiber/fiber/v2"
)

func participantTestApp(user *models.LocalAuthUser) *fiber.App {
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		if user != nil {
			c.Locals("localUser", *user)
		}
		return c.Next()
	})
	app.Put("/group/:code", ConfirmParticipation)
	app.Get("/group/:code", GetReservationProgress)
	app.Get("/reservations/:id/join-code", GetOwnerJoinCode)
	return app
}

func TestCancelledAndInvalidJoinCodesReturnIndistinguishable404(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	previous := database.DB
	database.DB = db
	defer func() { database.DB = previous }()

	for range 2 {
		mock.ExpectQuery("SELECT r.id,r.user_id,r.start_time").
			WillReturnRows(sqlmock.NewRows([]string{"id", "owner", "start", "minutes", "minimum"}))
		mock.ExpectQuery("(?s)SELECT r.id,r.status.*r.status IN \\('PENDING','CONFIRMED'\\)").
			WithArgs(sqlmock.AnyArg(), 3).
			WillReturnError(sql.ErrNoRows)
	}
	app := participantTestApp(&models.LocalAuthUser{ID: 3})
	var bodies [][]byte
	for _, path := range []string{"/group/codigo-cancelado", "/group/codigo-invalido"} {
		response, requestErr := app.Test(httptest.NewRequest(http.MethodGet, path, nil))
		if requestErr != nil || response.StatusCode != 404 {
			t.Fatalf("%s status=%d err=%v", path, response.StatusCode, requestErr)
		}
		body, readErr := io.ReadAll(response.Body)
		if readErr != nil {
			t.Fatal(readErr)
		}
		bodies = append(bodies, body)
	}
	if string(bodies[0]) != string(bodies[1]) || !strings.Contains(string(bodies[0]), "El código no existe o ya no está disponible.") {
		t.Fatalf("404 responses differ or leak state: %q / %q", bodies[0], bodies[1])
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestParticipationErrorStatusContract(t *testing.T) {
	cases := []struct {
		err  error
		want int
	}{
		{repositories.ErrParticipantIneligible, 403},
		{repositories.ErrInvalidJoinCode, 404},
		{repositories.ErrGroupCapacity, 409},
		{repositories.ErrOwnerCannotWithdraw, 409},
		{repositories.ErrParticipationDeadline, 410},
		{repositories.ErrParticipantConflict, 409},
		{sql.ErrConnDone, 500},
	}
	for _, item := range cases {
		if got := participationHTTPStatus(item.err); got != item.want {
			t.Fatalf("%v => %d, want %d", item.err, got, item.want)
		}
	}
}

func TestParticipationUnknownErrorDoesNotLeak(t *testing.T) {
	status, message := participationErrorResponse(errors.New("SECRET SQL ERROR"))
	if status != fiber.StatusInternalServerError || strings.Contains(message, "SECRET") {
		t.Fatalf("unknown participation error leaked: %d %q", status, message)
	}
}

func TestParticipantEndpointsReturn401WithoutAuthentication(t *testing.T) {
	response, err := participantTestApp(nil).Test(httptest.NewRequest(http.MethodPut, "/group/code", nil))
	if err != nil || response.StatusCode != 401 {
		t.Fatalf("status=%d err=%v", response.StatusCode, err)
	}
}

func TestExpiredParticipationReturns410(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	previous := database.DB
	database.DB = db
	defer func() { database.DB = previous }()
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT r.id,r.group_capacity_snapshot").
		WillReturnRows(sqlmock.NewRows([]string{"id", "capacity", "minimum", "target", "status", "start", "minutes", "deadline_minutes", "owner", "reason"}).
			AddRow(1, 20, 10, 12, "CANCELLED", time.Now(), 60, 90, 2, "CONFIRMATION_DEADLINE"))
	mock.ExpectRollback()
	user := models.LocalAuthUser{ID: 3}
	response, err := participantTestApp(&user).Test(httptest.NewRequest(http.MethodPut, "/group/code", nil))
	if err != nil || response.StatusCode != 410 {
		t.Fatalf("status=%d err=%v", response.StatusCode, err)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestJoinCodeUsesUniform404ForAdminWhoIsNotOwner(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	previous := database.DB
	database.DB = db
	defer func() { database.DB = previous }()
	mock.ExpectQuery("SELECT s.nonce,s.ciphertext,s.key_version").WithArgs(8, 99).
		WillReturnError(sql.ErrNoRows)
	user := models.LocalAuthUser{ID: 99, IsAdmin: true}
	response, _ := participantTestApp(&user).Test(httptest.NewRequest(http.MethodGet, "/reservations/8/join-code", nil))
	if response.StatusCode != 404 {
		t.Fatalf("status=%d", response.StatusCode)
	}
}

func TestJoinCodeIsUnavailableForTerminalReservations(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	previous := database.DB
	database.DB = db
	defer func() { database.DB = previous }()

	mock.ExpectQuery(`r.status IN\('PENDING','CONFIRMED'\)`).
		WithArgs(8, 3).
		WillReturnError(sql.ErrNoRows)
	user := models.LocalAuthUser{ID: 3}
	response, _ := participantTestApp(&user).Test(
		httptest.NewRequest(http.MethodGet, "/reservations/8/join-code", nil),
	)
	if response.StatusCode != 404 {
		t.Fatalf("status=%d", response.StatusCode)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
