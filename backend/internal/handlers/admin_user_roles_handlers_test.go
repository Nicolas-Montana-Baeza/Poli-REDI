package handlers

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"poli-redi-api/internal/repositories"
	"poli-redi-api/internal/services"
)

func TestAdminUserRoleError(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
		code int
	}{
		{"self", repositories.ErrUserRoleSelf, 403},
		{"not_found", repositories.ErrUserRoleNotFound, 404},
		{"already", repositories.ErrUserRoleUnchanged, 409},
		{"last", repositories.ErrUserRoleLastAdmin, 409},
		{"blocked", repositories.ErrUserRoleBlockedTarget, 409},
		{"concurrent", repositories.ErrUserRoleConflict, 409},
		{"reason", services.ErrUserRoleInvalidReason, 400},
		{"internal", errors.New("do not expose SQL"), 500},
	} {
		t.Run(tc.name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/", func(c *fiber.Ctx) error { return adminUserRoleError(c, tc.err) })
			response, err := app.Test(httptest.NewRequest("GET", "/", nil))
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()
			if response.StatusCode != tc.code {
				t.Fatalf("esperado HTTP %d, recibido %d", tc.code, response.StatusCode)
			}
		})
	}
}
