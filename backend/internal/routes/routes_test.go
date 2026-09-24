package routes

import (
	"testing"

	"github.com/gofiber/fiber/v2"
)

func routeSetForScope(t *testing.T, scope string) map[string]bool {
	t.Helper()

	t.Setenv("DEV_AUTH_ENABLED", "true")
	t.Setenv("MVP_SCOPE", scope)

	app := fiber.New()
	RegisterRoutes(app)

	routes := map[string]bool{}

	for _, route := range app.GetRoutes(true) {
		routes[route.Method+" "+route.Path] = true
	}

	return routes
}

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
		"GET /api/notifications",
		"PATCH /api/notifications/:id/read", "GET /api/workshops",
		"POST /api/workshops/:id/enroll", "GET /api/admin/reservation-policies",
		"POST /api/admin/reservation-policies", "PATCH /api/resources/:id/image",
	} {
		if routes[excluded] {
			t.Errorf("route outside MVP1 was registered: %s", excluded)
		}
	}
}

func TestMVP2DoesNotExposeMVP3Routes(t *testing.T) {
	routes := routeSetForScope(t, "mvp2")

	for _, excluded := range []string{
		"GET /api/notifications",
		"PATCH /api/notifications/:id/read",
		"GET /api/admin/reservation-policies",
		"POST /api/admin/reservation-policies",
		"PATCH /api/resources/:id/image",
	} {
		if routes[excluded] {
			t.Errorf("MVP3 route exposed in MVP2: %s", excluded)
		}
	}
}

func TestMVP3RouteSurface(t *testing.T) {
	routes := routeSetForScope(t, "mvp3")

	for _, required := range []string{
		// Herencia MVP2.
		"GET /api/reservations/join/:code",
		"POST /api/reservations/join/:code",
		"DELETE /api/reservations/join/:code",
		"GET /api/workshops",
		"GET /api/admin/scheduling-conflicts",

		// Superficie MVP3.
		"GET /api/notifications",
		"PATCH /api/notifications/:id/read",
		"GET /api/admin/reservation-policies",
		"POST /api/admin/reservation-policies",
		"PATCH /api/resources/:id/image",
	} {
		if !routes[required] {
			t.Errorf("missing MVP3 route %s", required)
		}
	}
}

func TestFullIncludesMVP3Routes(t *testing.T) {
	routes := routeSetForScope(t, "full")

	for _, required := range []string{
		"GET /api/notifications",
		"PATCH /api/notifications/:id/read",
		"GET /api/admin/reservation-policies",
		"POST /api/admin/reservation-policies",
		"PATCH /api/resources/:id/image",
	} {
		if !routes[required] {
			t.Errorf("FULL must include MVP3 route %s", required)
		}
	}
}
