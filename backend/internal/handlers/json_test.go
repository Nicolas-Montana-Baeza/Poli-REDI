package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"poli-redi-api/internal/models"

	"github.com/gofiber/fiber/v2"
)

func TestDecodeStrictJSONRejectsClientControlledReservationStatus(t *testing.T) {
	body := []byte(`{
		"resourceId": 1,
		"startTime": "2026-07-20T10:00:00",
		"durationMinutes": 60,
		"status": "CANCELLED"
	}`)

	var request models.CreateReservationRequest
	err := decodeStrictJSON(body, &request)

	if err == nil || !strings.Contains(err.Error(), "unknown field \"status\"") {
		t.Fatalf("decodeStrictJSON() error = %v, expected unknown status field", err)
	}
}

func TestDecodeStrictJSONAcceptsReservationContract(t *testing.T) {
	body := []byte(`{
		"resourceId": 1,
		"activityId": 2,
		"startTime": "2026-07-20T10:00:00",
		"durationMinutes": 60
	}`)

	var request models.CreateReservationRequest
	if err := decodeStrictJSON(body, &request); err != nil {
		t.Fatalf("decodeStrictJSON() error = %v", err)
	}
}

func TestCreateReservationRejectsStatusBeforeDatabaseAccess(t *testing.T) {
	app := fiber.New()
	app.Post("/reservations", func(c *fiber.Ctx) error {
		c.Locals("localUser", models.LocalAuthUser{ID: 1, IsAdmin: true})
		return CreateReservation(c)
	})

	body := `{
		"resourceId": 1,
		"startTime": "2026-07-20T10:00:00",
		"durationMinutes": 60,
		"status": "CANCELLED"
	}`
	request := httptest.NewRequest(
		http.MethodPost,
		"/reservations",
		strings.NewReader(body),
	)
	request.Header.Set("Content-Type", "application/json")

	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("status = %d, expected %d", response.StatusCode, fiber.StatusBadRequest)
	}
}
