package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"poli-redi-api/internal/models"

	"github.com/gofiber/fiber/v2"
)

func TestAdminRoutesHTTPAuthorization(t *testing.T) {
	endpoints := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/admin/availability-blocks"},
		{http.MethodPost, "/api/admin/availability-blocks"},
		{http.MethodPatch, "/api/admin/availability-blocks/1"},
		{http.MethodGet, "/api/admin/venues"},
		{http.MethodPost, "/api/admin/venues"},
		{http.MethodPatch, "/api/admin/venues/1"},
		{http.MethodPost, "/api/admin/resources"},
		{http.MethodPatch, "/api/resources/1"},
		{http.MethodPatch, "/api/resources/1/image"},
	}

	scenarios := []struct {
		name   string
		mount  func(*fiber.App)
		status int
	}{
		{
			name:   "sin_autenticacion",
			mount:  RegisterRoutes,
			status: fiber.StatusUnauthorized,
		},
		{
			name: "usuario_no_administrador",
			mount: func(app *fiber.App) {
				registerRoutes(app, func(c *fiber.Ctx) error {
					c.Locals("localUser", models.LocalAuthUser{
						IsAdmin: false,
					})
					return c.Next()
				})
			},
			status: fiber.StatusForbidden,
		},
	}

	for _, scope := range []string{"mvp3", "full"} {
		t.Run(scope, func(t *testing.T) {
			t.Setenv("MVP_SCOPE", scope)
			t.Setenv("DEV_AUTH_ENABLED", "true")

			for _, scenario := range scenarios {
				t.Run(scenario.name, func(t *testing.T) {
					for _, endpoint := range endpoints {
						t.Run(
							endpoint.method+" "+endpoint.path,
							func(t *testing.T) {
								app := fiber.New()
								scenario.mount(app)

								request := httptest.NewRequest(
									endpoint.method,
									endpoint.path,
									nil,
								)

								response, err := app.Test(request)
								if err != nil {
									t.Fatalf("Solicitud HTTP: %v", err)
								}
								defer response.Body.Close()

								if response.StatusCode != scenario.status {
									t.Errorf(
										"Esperado HTTP %d, obtenido %d",
										scenario.status,
										response.StatusCode,
									)
								}
							},
						)
					}
				})
			}
		})
	}
}
