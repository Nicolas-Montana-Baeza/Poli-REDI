package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"

	"github.com/gofiber/fiber/v2"
)

// Comprueba errores HTTP reales usando una conexion cerrada.
// No utiliza ni modifica bases de datos existentes.
func TestSEC005HTTPInternalErrorsAreSanitized(t *testing.T) {
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

	// Identidad controlada: no se habilita DEV_AUTH
	// ni se inicializa Microsoft Entra ID.
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("localUser", models.LocalAuthUser{
			ID:      1,
			IsAdmin: true,
		})
		return c.Next()
	})

	app.Get("/api/reservations", GetReservations)
	app.Get("/api/reservations/mine", GetMyReservations)
	app.Get("/api/availability/reservations", GetAvailabilityReservations)
	app.Get("/api/resources", GetResources)
	app.Get("/api/activities", GetActivities)
	app.Get("/api/notifications", GetNotifications)
	app.Post("/api/reservations", CreateReservation)
	app.Patch("/api/reservations/cancel", CancelReservation)

	start := time.Now().UTC().Add(48 * time.Hour).Format(time.RFC3339)

	tests := []struct {
		name    string
		method  string
		path    string
		body    string
		message string
	}{
		{
			"listar_reservas", http.MethodGet,
			"/api/reservations", "",
			"No se pudieron cargar las reservas",
		},
		{
			"reservas_propias", http.MethodGet,
			"/api/reservations/mine", "",
			"No se pudieron cargar tus reservas",
		},
		{
			"disponibilidad", http.MethodGet,
			"/api/availability/reservations", "",
			"No se pudo cargar la disponibilidad",
		},
		{
			"recursos", http.MethodGet,
			"/api/resources", "",
			"Error obteniendo recursos",
		},
		{
			"actividades", http.MethodGet,
			"/api/activities", "",
			"Error obteniendo actividades",
		},
		{
			"notificaciones", http.MethodGet,
			"/api/notifications", "",
			"Error obteniendo notificaciones",
		},
		{
			"crear_reserva", http.MethodPost,
			"/api/reservations",
			fmt.Sprintf(
				`{"resourceId":1,"startTime":%q,"durationMinutes":60}`,
				start,
			),
			"no se pudo crear la reserva",
		},
		{
			"cancelar_reserva", http.MethodPatch,
			"/api/reservations/cancel",
			`{"reservationId":1}`,
			"no se pudo cancelar la reserva",
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

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}

			if resp.StatusCode != fiber.StatusInternalServerError {
				t.Fatalf(
					"HTTP %d, esperado 500; respuesta: %s",
					resp.StatusCode,
					body,
				)
			}

			var result map[string]json.RawMessage
			if err := json.Unmarshal(body, &result); err != nil {
				t.Fatalf("JSON invalido: %v", err)
			}

			if len(result) != 1 {
				t.Fatalf("campos inesperados en respuesta: %s", body)
			}

			var message string
			if err := json.Unmarshal(result["error"], &message); err != nil {
				t.Fatalf("campo error invalido: %v", err)
			}

			if message != tc.message {
				t.Fatalf(
					"mensaje %q, esperado %q",
					message,
					tc.message,
				)
			}
		})
	}
}
