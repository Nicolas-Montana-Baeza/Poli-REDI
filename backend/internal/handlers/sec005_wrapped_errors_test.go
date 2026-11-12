package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"poli-redi-api/internal/repositories"
	"poli-redi-api/internal/services"

	"github.com/gofiber/fiber/v2"
)

func TestSEC005WrappedSentinelHTTP(t *testing.T) {
	type responder func(*fiber.Ctx, error) error

	wrapResponse := func(
		classifier func(error) (int, string),
	) responder {
		return func(c *fiber.Ctx, err error) error {
			status, message := classifier(err)
			return c.Status(status).JSON(fiber.Map{
				"error": message,
			})
		}
	}

	tests := []struct {
		name    string
		known   error
		status  int
		respond responder
	}{
		{
			"admin_access",
			repositories.ErrUserAccessNotFound,
			404,
			adminUserAccessError,
		},
		{
			"admin_role",
			repositories.ErrUserRoleLastAdmin,
			409,
			adminUserRoleError,
		},
		{
			"institutional_activity",
			services.ErrInstitutionalUnauthorized,
			403,
			institutionalActivityHTTPError,
		},
		{
			"institutional_unit",
			services.ErrInstitutionalUnitNotFound,
			404,
			institutionalUnitHTTPError,
		},
		{
			"institutional_occurrence",
			repositories.ErrInstitutionalOccurrenceBlocked,
			409,
			institutionalOccurrenceHTTPError,
		},
		{
			"institutional_workshop",
			repositories.ErrInstitutionalWorkshopFull,
			409,
			wrapResponse(institutionalWorkshopErrorResponse),
		},
		{
			"scheduling_conflict",
			services.ErrSchedulingConflictInvalidResolution,
			400,
			wrapResponse(schedulingConflictErrorResponse),
		},
		{
			"group_participation",
			services.ErrInvalidJoinCode,
			404,
			participantErrorResponse,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			wrapped := fmt.Errorf(
				"SQLSTATE XX999 SECRET_DATABASE_URL: %w",
				tc.known,
			)

			if !errors.Is(wrapped, tc.known) {
				t.Fatal("el error de dominio no se conserva")
			}

			app := fiber.New()
			app.Get("/test", func(c *fiber.Ctx) error {
				return tc.respond(c, wrapped)
			})

			req := httptest.NewRequest(
				http.MethodGet, "/test", nil,
			)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			var payload map[string]json.RawMessage
			if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
				t.Fatal(err)
			}

			if resp.StatusCode != tc.status {
				t.Fatalf(
					"HTTP %d, esperado %d",
					resp.StatusCode, tc.status,
				)
			}

			if len(payload) != 1 {
				t.Fatalf("campos inesperados: %v", payload)
			}

			var message string
			if err := json.Unmarshal(
				payload["error"], &message,
			); err != nil {
				t.Fatal(err)
			}

			if message != tc.known.Error() {
				t.Fatalf(
					"mensaje publico %q, esperado %q",
					message, tc.known.Error(),
				)
			}

			if strings.Contains(message, "SECRET_DATABASE_URL") ||
				strings.Contains(message, "SQLSTATE") {
				t.Fatalf(
					"filtracion de error interno: %q",
					message,
				)
			}
		})
	}
}
