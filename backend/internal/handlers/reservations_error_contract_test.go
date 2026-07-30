package handlers

import (
	"errors"
	"testing"

	"poli-redi-api/internal/repositories"

	"github.com/gofiber/fiber/v2"
)

func TestCreateReservationErrorResponseDoesNotExposeUnknownInternals(t *testing.T) {
	status, message := createReservationErrorResponse(errors.New("SECRET SQL TABLE/CONSTRAINT DETAIL"))
	if status != fiber.StatusInternalServerError || message != "No se pudo crear la reserva" {
		t.Fatalf("unknown error leaked: status=%d message=%q", status, message)
	}
}

func TestCreateReservationErrorResponseKeepsPublicOverlapContract(t *testing.T) {
	status, message := createReservationErrorResponse(repositories.ErrParticipantConflict)
	if status != fiber.StatusConflict || message != repositories.ErrParticipantConflict.Error() {
		t.Fatalf("overlap response = %d %q", status, message)
	}
}

func TestCancelReservationErrorResponseDoesNotExposeUnknownInternals(t *testing.T) {
	status, message := cancelReservationErrorResponse(errors.New("SECRET DRIVER DETAIL"))
	if status != fiber.StatusInternalServerError || message != "No se pudo cancelar la reserva" {
		t.Fatalf("unknown cancellation leaked: status=%d message=%q", status, message)
	}
}
