package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"poli-redi-api/internal/models"

	"github.com/gofiber/fiber/v2"
)

func TestRequireAdminHTTP(t *testing.T) {
	tests := []struct {
		name       string
		hasUser    bool
		isAdmin    bool
		wantStatus int
	}{
		{
			name:       "sin autenticar",
			wantStatus: fiber.StatusUnauthorized,
		},
		{
			name:       "usuario sin permisos",
			hasUser:    true,
			wantStatus: fiber.StatusForbidden,
		},
		{
			name:       "administrador autorizado",
			hasUser:    true,
			isAdmin:    true,
			wantStatus: fiber.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()

			app.Use(func(c *fiber.Ctx) error {
				if tt.hasUser {
					c.Locals("localUser", models.LocalAuthUser{
						IsAdmin: tt.isAdmin,
					})
				}

				return c.Next()
			})

			app.Get(
				"/admin/test",
				RequireAdmin(),
				func(c *fiber.Ctx) error {
					return c.SendStatus(fiber.StatusOK)
				},
			)

			request := httptest.NewRequest(
				http.MethodGet,
				"/admin/test",
				nil,
			)

			response, err := app.Test(request)
			if err != nil {
				t.Fatalf("Error HTTP: %v", err)
			}
			defer response.Body.Close()

			if response.StatusCode != tt.wantStatus {
				t.Errorf(
					"HTTP esperado %d, obtenido %d",
					tt.wantStatus,
					response.StatusCode,
				)
			}
		})
	}
}
