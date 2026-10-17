package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestGetResourcesRejectsInvalidFilters(t *testing.T) {
	app := fiber.New()
	app.Get("/resources", GetResources)

	cases := []string{
		"/resources?venueId=abc",
		"/resources?venueId=0",
		"/resources?venueId=-1",
		"/resources?isActive=invalid",
		"/resources?isActive=1",
		"/resources?type=%20",
	}

	for _, path := range cases {
		t.Run(path, func(t *testing.T) {
			request := httptest.NewRequest("GET", path, nil)
			response, err := app.Test(request)
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()

			if response.StatusCode != fiber.StatusBadRequest {
				t.Fatalf(
					"GET %s: esperado 400, obtenido %d",
					path,
					response.StatusCode,
				)
			}
		})
	}
}
