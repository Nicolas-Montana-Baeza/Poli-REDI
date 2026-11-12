package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"

	"github.com/gofiber/fiber/v2"
)

func TestSEC005UpdateRUTHTTP(t *testing.T) {
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

	tests := []struct {
		name       string
		admin      bool
		body       string
		wantStatus int
		wantError  string
	}{
		{
			name:       "usuario_error_postgresql",
			body:       `{"rut":"12.345.678-5"}`,
			wantStatus: 500,
			wantError:  "no se pudo actualizar el RUT",
		},
		{
			name:       "admin_error_postgresql",
			admin:      true,
			body:       `{"rut":""}`,
			wantStatus: 500,
			wantError:  "no se pudo actualizar el RUT",
		},
		{
			name:       "rut_invalido_sigue_siendo_400",
			body:       `{"rut":"invalido"}`,
			wantStatus: 400,
			wantError:  "RUT inválido",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			app := fiber.New()

			app.Use(func(c *fiber.Ctx) error {
				c.Locals("localUser", models.LocalAuthUser{
					ID:      42,
					IsAdmin: tc.admin,
				})
				return c.Next()
			})

			app.Patch("/api/me/rut", UpdateMeRUT)

			req := httptest.NewRequest(
				http.MethodPatch,
				"/api/me/rut",
				strings.NewReader(tc.body),
			)
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			var payload map[string]string
			if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
				t.Fatal(err)
			}

			if resp.StatusCode != tc.wantStatus {
				t.Fatalf(
					"HTTP %d, esperado %d",
					resp.StatusCode, tc.wantStatus,
				)
			}

			if payload["error"] != tc.wantError {
				t.Fatalf("respuesta inesperada: %v", payload)
			}

			if tc.wantStatus == 500 {
				if payload["code"] != "INTERNAL_ERROR" ||
					len(payload) != 2 {
					t.Fatalf("respuesta 500 insegura: %v", payload)
				}
			}
		})
	}
}
