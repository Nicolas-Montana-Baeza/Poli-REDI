package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"poli-redi-api/internal/models"

	"github.com/gofiber/fiber/v2"
)

func sec005AssertAuthResponse(
	t *testing.T,
	app *fiber.App,
	wantStatus int,
	wantCode string,
	wantMessage string,
) {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/secure", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != wantStatus {
		t.Fatalf("HTTP %d, esperado %d", resp.StatusCode, wantStatus)
	}

	var payload map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}

	if len(payload) != 2 ||
		payload["code"] != wantCode ||
		payload["error"] != wantMessage {
		t.Fatalf("contrato HTTP inesperado: %v", payload)
	}

	if _, exists := payload["detail"]; exists {
		t.Fatal("la respuesta incluye detail")
	}
}

func TestSEC005DevAuthMissingHeaderContract(t *testing.T) {
	app := fiber.New()

	app.Get("/secure", requireDevAuth(), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	sec005AssertAuthResponse(
		t,
		app,
		401,
		"UNAUTHORIZED",
		"falta iniciar sesión local de prueba",
	)
}

func TestSEC005RequireAdminPublicContract(t *testing.T) {
	tests := []struct {
		name          string
		authenticated bool
		admin         bool
		status        int
		code          string
		message       string
	}{
		{
			name:    "sin_autenticar",
			status:  401,
			code:    "UNAUTHORIZED",
			message: "usuario no autenticado",
		},
		{
			name:          "sin_permisos",
			authenticated: true,
			status:        403,
			code:          "FORBIDDEN",
			message:       "no tienes permisos para acceder a esta sección",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			app := fiber.New()

			app.Use(func(c *fiber.Ctx) error {
				if tc.authenticated {
					c.Locals("localUser", models.LocalAuthUser{
						ID:      42,
						IsAdmin: tc.admin,
					})
				}
				return c.Next()
			})

			app.Get("/secure", RequireAdmin(), func(c *fiber.Ctx) error {
				return c.SendStatus(fiber.StatusOK)
			})

			sec005AssertAuthResponse(
				t, app, tc.status, tc.code, tc.message,
			)
		})
	}
}
