package routes

import "testing"

func TestCreateResourceRouteScope(t *testing.T) {
	for _, scope := range []string{"mvp1", "mvp2"} {
		routes := routeSetForScope(t, scope)
		if routes["POST /api/admin/resources"] {
			t.Errorf("%s expone creacion de recursos", scope)
		}
	}

	for _, scope := range []string{"mvp3", "full"} {
		routes := routeSetForScope(t, scope)
		if !routes["POST /api/admin/resources"] {
			t.Errorf("%s no registra creacion de recursos", scope)
		}
	}
}
