package handlers

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"poli-redi-api/internal/repositories"
	"poli-redi-api/internal/services"
)

func TestAdminUserAccessHTTPError(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
		code int
	}{
		{"self", repositories.ErrUserAccessSelf, fiber.StatusForbidden},
		{"not_found", repositories.ErrUserAccessNotFound, fiber.StatusNotFound},
		{"already", repositories.ErrUserAccessUnchanged, fiber.StatusConflict},
		{"last_admin", repositories.ErrUserAccessLastAdmin, fiber.StatusConflict},
		{"concurrent", repositories.ErrUserAccessConflict, fiber.StatusConflict},
		{"invalid_reason", services.ErrUserAccessInvalidReason, fiber.StatusBadRequest},
		{"internal_sanitized", errors.New("sensitive SQL detail"), fiber.StatusInternalServerError},
	} {
		t.Run(tc.name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/", func(c *fiber.Ctx) error { return adminUserAccessError(c, tc.err) })
			response, err := app.Test(httptest.NewRequest("GET", "/", nil))
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()
			if response.StatusCode != tc.code {
				t.Fatalf("status esperado %d, obtenido %d", tc.code, response.StatusCode)
			}
		})
	}
}
