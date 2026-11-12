package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"

	"github.com/gofiber/fiber/v2"
)

func TestSEC005DynamicHTTPContracts(t *testing.T) {
	previousDB := database.DB

	db, err := sql.Open(
		"pgx",
		"postgres://sec005:sec005@127.0.0.1:1/sec005?sslmode=disable",
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

	app.Get("/test/workshops", GetInstitutionalWorkshops)
	app.Get("/test/workshops/:id", GetInstitutionalWorkshop)
	app.Post("/test/workshops/:id/enroll", EnrollInInstitutionalWorkshop)
	app.Delete("/test/workshops/:id/enroll", LeaveInstitutionalWorkshop)

	app.Get("/test/conflicts", GetSchedulingConflicts)
	app.Get("/test/conflicts/:id", GetSchedulingConflict)
	app.Patch(
		"/test/conflicts/:id/items/:itemId",
		PatchSchedulingConflictItem,
	)

	app.Post("/test/policies", PublishReservationPolicy)

	tests := []struct {
		name    string
		method  string
		path    string
		body    string
		status  int
		code    string
		message string
	}{
		{
			"listar_talleres", "GET", "/test/workshops", "",
			500, "INTERNAL_ERROR", "no se pudo procesar el taller",
		},
		{
			"detalle_taller", "GET", "/test/workshops/1", "",
			500, "INTERNAL_ERROR", "no se pudo procesar el taller",
		},
		{
			"inscribir_taller", "POST", "/test/workshops/1/enroll", "",
			500, "INTERNAL_ERROR", "no se pudo procesar el taller",
		},
		{
			"retirar_taller", "DELETE", "/test/workshops/1/enroll", "",
			500, "INTERNAL_ERROR", "no se pudo procesar el taller",
		},
		{
			"listar_conflictos", "GET", "/test/conflicts", "",
			500, "INTERNAL_ERROR",
			"no se pudo procesar el conflicto de programación",
		},
		{
			"detalle_conflicto", "GET", "/test/conflicts/1", "",
			500, "INTERNAL_ERROR",
			"no se pudo procesar el conflicto de programación",
		},
		{
			"resolver_conflicto", "PATCH",
			"/test/conflicts/1/items/1",
			`{"resolution":"KEEP","resolutionNote":"prueba SEC005"}`,
			500, "INTERNAL_ERROR",
			"no se pudo procesar el conflicto de programación",
		},
		{
			"validacion_politica", "POST", "/test/policies",
			`{}`,
			400, "VALIDATION_ERROR",
			"expectedPolicyId es obligatorio y debe ser positivo",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(
				tc.method,
				tc.path,
				strings.NewReader(tc.body),
			)
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tc.status {
				t.Fatalf(
					"HTTP %d, esperado %d",
					resp.StatusCode, tc.status,
				)
			}

			var payload map[string]string
			if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
				t.Fatal(err)
			}

			if len(payload) != 2 ||
				payload["code"] != tc.code ||
				payload["error"] != tc.message {
				t.Fatalf(
					"contrato HTTP inesperado: %v",
					payload,
				)
			}

			if strings.Contains(
				payload["error"], "SQLSTATE",
			) {
				t.Fatalf(
					"detalle SQL expuesto: %v",
					payload,
				)
			}
		})
	}
}
