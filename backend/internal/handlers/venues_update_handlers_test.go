package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestUpdateAdminVenueRejectsInvalidRequests(t *testing.T) {
	tests := []struct {
		name string
		id   string
		body string
	}{
		{"ID invalido", "0", `{"name":"Nueva sede"}`},
		{"Sin campos", "1", `{}`},
		{"Nombre vacio", "1", `{"name":"  "}`},
		{"Nombre nulo", "1", `{"name":null}`},
		{"Pais vacio", "1", `{"country":""}`},
		{"Estado nulo", "1", `{"isActive":null}`},
		{"Latitud invalida", "1", `{"latitude":91}`},
		{"Longitud invalida", "1", `{"longitude":-181}`},
		{"Campo desconocido", "1", `{"unknown":true}`},
		{"Direccion numerica", "1", `{"addressLine":123}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Patch("/venues/:id", UpdateAdminVenue)

			req := httptest.NewRequest(
				http.MethodPatch,
				"/venues/"+tt.id,
				strings.NewReader(tt.body),
			)
			req.Header.Set(
				"Content-Type",
				"application/json",
			)

			resp, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusBadRequest {
				t.Fatalf(
					"esperado 400, obtenido %d",
					resp.StatusCode,
				)
			}
		})
	}
}
