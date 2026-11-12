package handlers

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"poli-redi-api/internal/models"
	"poli-redi-api/internal/repositories"
	"poli-redi-api/internal/services"

	"github.com/gofiber/fiber/v2"
)

func TestSEC005DomainSentinelMessages(t *testing.T) {
	tests := []struct {
		name  string
		known error
	}{
		{"recurso", repositories.ErrResourceDuplicate},
		{"sede", repositories.ErrVenueDuplicate},
		{"bloqueo", repositories.ErrBlockOverlap},
		{"reserva", services.ErrReservationNotFound},
		{"capacidad_grupal",
			repositories.ErrGroupResourceCapacityInvalid},
		{"recurso_no_disponible",
			repositories.ErrBlockResourceUnavailable},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			wrapped := fmt.Errorf(
				"SQLSTATE XX999 SECRET_DATABASE_URL: %w",
				tc.known,
			)

			got := sec005KnownSentinelMessage(
				wrapped,
				tc.known,
			)

			if got != tc.known.Error() {
				t.Fatalf("mensaje incorrecto: %q", got)
			}

			if strings.Contains(got, "SECRET_DATABASE_URL") ||
				strings.Contains(got, "SQLSTATE") {
				t.Fatalf("filtracion tecnica: %q", got)
			}
		})
	}
}

func TestSEC005ReservationPolicyWrappedErrors(t *testing.T) {
	_, _, validation := services.PublishReservationPolicy(
		models.PublishReservationPolicyRequest{},
		1,
		"",
	)

	if validation == nil {
		t.Fatal("se esperaba error de validacion")
	}

	tests := []struct {
		name   string
		err    error
		status int
		want   string
	}{
		{
			"validacion_envuelta",
			validation,
			fiber.StatusBadRequest,
			validation.Error(),
		},
		{
			"conflicto_envuelto",
			services.ErrReservationPolicyConflict,
			fiber.StatusConflict,
			services.ErrReservationPolicyConflict.Error(),
		},
		{
			"concurrencia_envuelta",
			services.ErrReservationPolicyConcurrent,
			fiber.StatusConflict,
			services.ErrReservationPolicyConcurrent.Error(),
		},
		{
			"version_obsoleta_envuelta",
			services.ErrReservationPolicyStale,
			fiber.StatusConflict,
			services.ErrReservationPolicyStale.Error(),
		},
		{
			"error_desconocido",
			errors.New("database unavailable"),
			fiber.StatusInternalServerError,
			"no se pudo publicar la politica de reservas",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			wrapped := fmt.Errorf(
				"SECRET_DATABASE_URL: %w",
				tc.err,
			)

			status, message := reservationPolicyErrorResponse(
				wrapped,
			)

			if status != tc.status || message != tc.want {
				t.Fatalf(
					"obtenido (%d, %q), esperado (%d, %q)",
					status, message, tc.status, tc.want,
				)
			}

			if strings.Contains(message, "SECRET_DATABASE_URL") {
				t.Fatalf("filtracion tecnica: %q", message)
			}
		})
	}
}
