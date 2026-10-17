package routes

import "testing"

func TestAdminVenuesRouteScope(t *testing.T) {
	for _, scope := range []string{"mvp1", "mvp2"} {
		routes := routeSetForScope(t, scope)

		for _, route := range []string{
			"GET /api/admin/venues",
			"POST /api/admin/venues",
			"PATCH /api/admin/venues/:id",
		} {
			if routes[route] {
				t.Errorf("%s expone %s", scope, route)
			}
		}
	}

	for _, scope := range []string{"mvp3", "full"} {
		routes := routeSetForScope(t, scope)

		for _, route := range []string{
			"GET /api/admin/venues",
			"POST /api/admin/venues",
			"PATCH /api/admin/venues/:id",
		} {
			if !routes[route] {
				t.Errorf("%s no registra %s", scope, route)
			}
		}
	}
}
