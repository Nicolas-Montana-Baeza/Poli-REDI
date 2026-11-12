package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"poli-redi-api/internal/repositories"
	"poli-redi-api/internal/services"

	"github.com/gofiber/fiber/v2"
)

func TestCreateReservationPublicError(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		status  int
		message string
	}{
		{
			name:    "SQL no encontrado",
			err:     sql.ErrNoRows,
			status:  fiber.StatusBadRequest,
			message: "la instalación seleccionada no existe",
		},
		{
			name:    "validacion de horario",
			err:     errors.New("la hora de inicio debe usar intervalos de 15 minutos"),
			status:  fiber.StatusBadRequest,
			message: "la hora de inicio debe usar intervalos de 15 minutos",
		},
		{
			name: "frecuencia",
			err: repositories.RequestFrequencyError{
				NextDate: time.Date(2026, 11, 18, 0, 0, 0, 0, time.UTC),
			},
			status:  fiber.StatusBadRequest,
			message: "ya tienes una solicitud vigente; proxima fecha permitida: 2026-11-18",
		},
		{
			name:    "error SQL desconocido",
			err:     errors.New("SQLSTATE XX999 connection password=secreto"),
			status:  fiber.StatusInternalServerError,
			message: "no se pudo crear la reserva",
		},
		{
			name:    "error interno envuelto",
			err:     fmt.Errorf("repository: %w", errors.New("tabla privada users")),
			status:  fiber.StatusInternalServerError,
			message: "no se pudo crear la reserva",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			status, message := createReservationPublicError(tc.err)
			if status != tc.status || message != tc.message {
				t.Fatalf(
					"got (%d, %q), want (%d, %q)",
					status, message, tc.status, tc.message,
				)
			}
		})
	}
}

func TestCancelReservationPublicError(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
	}{
		{
			name:   "reserva inexistente",
			err:    services.ErrReservationNotFound,
			status: fiber.StatusNotFound,
		},
		{
			name:   "sin permisos",
			err:    repositories.ErrReservationForbidden,
			status: fiber.StatusForbidden,
		},
		{
			name:   "cancelacion duplicada",
			err:    repositories.ErrReservationNotCancellable,
			status: fiber.StatusBadRequest,
		},
		{
			name:   "fallo inesperado",
			err:    errors.New("connection password=secreto"),
			status: fiber.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			status, message := cancelReservationPublicError(tc.err)

			if status != tc.status {
				t.Fatalf("got %d, want %d", status, tc.status)
			}

			if strings.Contains(message, "secreto") ||
				strings.Contains(message, "connection") {
				t.Fatalf("se filtro un error interno: %q", message)
			}
		})
	}
}
