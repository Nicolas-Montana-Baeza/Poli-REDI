package routes

import (
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestMVP1RouteAllowlist(t *testing.T) {
	t.Setenv("DEV_AUTH_ENABLED", "true")
	t.Setenv("MVP_SCOPE", "mvp1")
	app := fiber.New()
	RegisterRoutes(app)

	routes := map[string]bool{}
	for _, route := range app.GetRoutes(true) {
		routes[route.Method+" "+route.Path] = true
	}

	for _, allowed := range []string{
		"GET /api/health", "GET /api/me", "PATCH /api/me/rut",
		"GET /api/resources", "GET /api/activities",
		"GET /api/reservation-policy/current", "GET /api/availability/reservations",
		"GET /api/reservations/mine", "GET /api/reservations/:id",
		"POST /api/reservations", "PATCH /api/reservations/cancel",
		"GET /api/users", "GET /api/reservations",
	} {
		if !routes[allowed] {
			t.Errorf("missing MVP1 route %s", allowed)
		}
	}

	for _, excluded := range []string{
		"GET /api/notifications", "GET /api/workshops",
		"POST /api/workshops/:id/enroll", "GET /api/admin/reservation-policies",
		"POST /api/admin/reservation-policies", "PATCH /api/resources/:id/image",
	} {
		if routes[excluded] {
			t.Errorf("route outside MVP1 was registered: %s", excluded)
		}
	}
}
