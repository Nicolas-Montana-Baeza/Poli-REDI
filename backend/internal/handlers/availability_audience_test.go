package handlers

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"poli-redi-api/internal/models"

	"github.com/gofiber/fiber/v2"
)

func availabilityResponseForUser(
	t *testing.T,
	user models.LocalAuthUser,
	items []models.AvailabilityItem,
) []models.AvailabilityItem {
	t.Helper()

	previous := getAvailabilityItems
	getAvailabilityItems = func() ([]models.AvailabilityItem, error) {
		return items, nil
	}
	t.Cleanup(func() {
		getAvailabilityItems = previous
	})

	app := fiber.New()
	app.Get("/availability", func(c *fiber.Ctx) error {
		c.Locals("localUser", user)
		return GetAvailabilityReservations(c)
	})

	response, err := app.Test(httptest.NewRequest("GET", "/availability", nil))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()

	if response.StatusCode != fiber.StatusOK {
		body, _ := io.ReadAll(response.Body)
		t.Fatalf("status = %d, body = %s", response.StatusCode, string(body))
	}

	var payload []models.AvailabilityItem
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}

	return payload
}

func TestGetAvailabilityReservationsPreservesOwnAndRedactsForeign(t *testing.T) {
	target := 10
	items := []models.AvailabilityItem{
		{
			ID:               1,
			AvailabilityKind: models.AvailabilityKindReservation,
			UserID:           7,
			Title:            "Basquetbol",
			UserEmail:        "self@example.test",
		},
		{
			ID:                  2,
			AvailabilityKind:    models.AvailabilityKindGroupReservation,
			UserID:              99,
			Title:               "Actividad privada",
			UserEmail:           "owner@example.test",
			TargetParticipants:  &target,
			ParticipantCount:    4,
			MinimumParticipants: 10,
		},
	}

	payload := availabilityResponseForUser(t, models.LocalAuthUser{ID: 7}, items)

	if payload[0].UserID != 7 || payload[0].Title != "Basquetbol" || payload[0].UserEmail != "" {
		t.Fatalf("own availability item = %+v", payload[0])
	}
	if payload[1].UserID != 0 || payload[1].Title != "Reserva" ||
		payload[1].UserEmail != "" || payload[1].TargetParticipants != nil ||
		payload[1].ParticipantCount != 0 || payload[1].MinimumParticipants != 0 {
		t.Fatalf("foreign availability item = %+v", payload[1])
	}
	if payload[1].AvailabilityKind != models.AvailabilityKindGroupReservation {
		t.Fatalf("foreign safe kind = %q", payload[1].AvailabilityKind)
	}
}

func TestGetAvailabilityReservationsKeepsAdministratorOperationalDetail(t *testing.T) {
	target := 10
	items := []models.AvailabilityItem{{
		ID:                  2,
		AvailabilityKind:    models.AvailabilityKindGroupReservation,
		UserID:              99,
		Title:               "Basquetbol",
		UserEmail:           "owner@example.test",
		TargetParticipants:  &target,
		ParticipantCount:    4,
		MinimumParticipants: 10,
	}}

	payload := availabilityResponseForUser(
		t,
		models.LocalAuthUser{ID: 1, IsAdmin: true},
		items,
	)

	if payload[0].UserID != 99 || payload[0].Title != "Basquetbol" ||
		payload[0].UserEmail != "owner@example.test" ||
		payload[0].TargetParticipants == nil || payload[0].ParticipantCount != 4 {
		t.Fatalf("administrator detail changed: %+v", payload[0])
	}
}

func TestGetAvailabilityReservationsKeepsScheduledTypeButHidesCreator(t *testing.T) {
	items := []models.AvailabilityItem{{
		ID:                  3,
		AvailabilityKind:    models.AvailabilityKindScheduled,
		UserID:              42,
		Title:               "Entrenamiento selección universitaria",
		UserEmail:           "admin@example.test",
		IsScheduledActivity: true,
		ActivityType:        "TRAINING",
	}}

	payload := availabilityResponseForUser(t, models.LocalAuthUser{ID: 7}, items)
	item := payload[0]

	if item.Title != "Actividad institucional" ||
		item.ActivityType != "TRAINING" ||
		item.AvailabilityKind != models.AvailabilityKindScheduled ||
		item.UserID != 0 || item.UserEmail != "" {
		t.Fatalf("scheduled activity response = %+v", item)
	}
}
