package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestWorkshopWithdrawalRouteIsProtectedAndRegistered(t *testing.T) {
	t.Setenv("DEV_AUTH_ENABLED", "true")
	app := fiber.New()
	RegisterRoutes(app)

	response, err := app.Test(httptest.NewRequest(
		http.MethodDelete,
		"/api/workshops/9/enrollment",
		nil,
	))
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("status=%d, want protected route status %d", response.StatusCode, fiber.StatusUnauthorized)
	}
}
