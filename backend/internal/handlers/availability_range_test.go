package handlers

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"poli-redi-api/internal/businessclock"
	"poli-redi-api/internal/models"

	"github.com/gofiber/fiber/v2"
)

func TestGetAvailabilityReservationsUsesInclusiveBusinessDateRange(t *testing.T) {
	previous := getAvailabilityItemsForRange
	t.Cleanup(func() { getAvailabilityItemsForRange = previous })

	var gotFrom, gotToExclusive time.Time
	var gotUserID int
	getAvailabilityItemsForRange = func(from, toExclusive time.Time, userID int) ([]models.AvailabilityItem, error) {
		gotFrom, gotToExclusive, gotUserID = from, toExclusive, userID
		return []models.AvailabilityItem{}, nil
	}

	app := fiber.New()
	app.Get("/availability", func(c *fiber.Ctx) error {
		c.Locals("localUser", models.LocalAuthUser{ID: 77})
		return GetAvailabilityReservations(c)
	})
	response, err := app.Test(httptest.NewRequest(
		"GET", "/availability?from=2026-08-04&to=2026-08-05", nil,
	))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d", response.StatusCode)
	}
	if gotUserID != 77 || gotFrom.Location().String() != businessclock.Location().String() ||
		gotFrom.Format("2006-01-02 15:04") != "2026-08-04 00:00" ||
		gotToExclusive.Format("2006-01-02 15:04") != "2026-08-06 00:00" {
		t.Fatalf("range/user = %v .. %v, user %d", gotFrom, gotToExclusive, gotUserID)
	}
}

func TestGetAvailabilityReservationsRejectsInvalidRangesWithStableCodes(t *testing.T) {
	tests := []struct {
		query string
		code  string
	}{
		{"from=2026-08-04", "AVAILABILITY_RANGE_REQUIRED"},
		{"from=x&to=2026-08-04", "AVAILABILITY_RANGE_INVALID"},
		{"from=2026-08-05&to=2026-08-04", "AVAILABILITY_RANGE_INVALID"},
		{"from=2026-08-01&to=2026-09-01", "AVAILABILITY_RANGE_TOO_LARGE"},
	}

	for _, test := range tests {
		t.Run(test.code+test.query, func(t *testing.T) {
			app := fiber.New()
			app.Get("/availability", func(c *fiber.Ctx) error {
				c.Locals("localUser", models.LocalAuthUser{ID: 7})
				return GetAvailabilityReservations(c)
			})
			response, err := app.Test(httptest.NewRequest("GET", "/availability?"+test.query, nil))
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()
			if response.StatusCode != fiber.StatusBadRequest {
				t.Fatalf("status = %d", response.StatusCode)
			}
			var payload struct {
				Code string `json:"code"`
			}
			if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
				t.Fatal(err)
			}
			if payload.Code != test.code {
				t.Fatalf("code = %q, want %q", payload.Code, test.code)
			}
		})
	}
}

func TestParseAvailabilityRangeAllowsExactly31CalendarDaysAcrossDST(t *testing.T) {
	_, toExclusive, code, _ := parseAvailabilityRange("2026-09-01", "2026-10-01")
	if code != "" || toExclusive.Format("2006-01-02") != "2026-10-02" {
		t.Fatalf("31-day range rejected: code=%q to=%v", code, toExclusive)
	}
}
