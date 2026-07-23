package handlers

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gofiber/fiber/v2"
	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"
	"poli-redi-api/internal/repositories"
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
	app.Get("/reservations/:id/join-code", GetOwnerJoinCode)
	return app
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
	}
	for _, item := range cases {
		if got := participationHTTPStatus(item.err); got != item.want {
			t.Fatalf("%v => %d, want %d", item.err, got, item.want)
		}
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
		WillReturnRows(sqlmock.NewRows([]string{"id", "capacity", "minimum", "target", "status", "start", "minutes", "owner", "reason"}).
			AddRow(1, 20, 10, 12, "CANCELLED", time.Now(), 60, 2, "CONFIRMATION_DEADLINE"))
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
