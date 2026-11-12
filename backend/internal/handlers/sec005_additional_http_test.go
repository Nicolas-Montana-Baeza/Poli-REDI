package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"

	"github.com/gofiber/fiber/v2"
)

func TestSEC005AdditionalHTTP500Contract(t *testing.T) {
	previousDB := database.DB

	db, err := sql.Open(
		"pgx",
		"postgres://test:test@127.0.0.1:1/test?sslmode=disable",
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	database.DB = db
	t.Cleanup(func() {
		database.DB = previousDB
	})

	app := fiber.New()

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("localUser", models.LocalAuthUser{
			ID:      42,
			IsAdmin: true,
		})
		return c.Next()
	})

	app.Get("/test/users", GetUsers)
	app.Get("/test/venues", GetAdminVenues)
	app.Get("/test/policies/current", GetCurrentReservationPolicy)
	app.Get("/test/policies/history", GetReservationPolicyHistory)
	app.Get("/test/blocks", GetAdminAvailabilityBlocks)

	tests := []struct {
		name    string
		path    string
		message string
	}{
		{
			"usuarios",
			"/test/users",
			"Error obteniendo usuarios",
		},
		{
			"sedes",
			"/test/venues",
			"no se pudieron cargar las sedes",
		},
		{
			"politica_actual",
			"/test/policies/current",
			"no se pudo cargar la politica de reservas",
		},
		{
			"historial_politicas",
			"/test/policies/history",
			"no se pudo cargar el historial de politicas",
		},
		{
			"bloqueos_administrativos",
			"/test/blocks?from=2026-11-12&to=2026-11-13",
			"no se pudieron consultar los bloqueos",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(
				http.MethodGet, tc.path, nil,
			)

			resp, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != 500 {
				t.Fatalf(
					"HTTP %d, esperado 500",
					resp.StatusCode,
				)
			}

			var payload map[string]string
			if err := json.NewDecoder(
				resp.Body,
			).Decode(&payload); err != nil {
				t.Fatal(err)
			}

			if len(payload) != 2 ||
				payload["error"] != tc.message ||
				payload["code"] != "INTERNAL_ERROR" {
				t.Fatalf(
					"contrato HTTP inesperado: %v",
					payload,
				)
			}
		})
	}
}
