package handlers

import (
	"errors"
	"testing"

	"poli-redi-api/internal/models"
	"poli-redi-api/internal/services"
)

func TestReservationPolicyErrorResponse(t *testing.T) {
	_, _, validation := services.PublishReservationPolicy(models.PublishReservationPolicyRequest{}, 1, "")
	tests := []struct {
		err    error
		status int
	}{
		{validation, 400},
		{services.ErrReservationPolicyConflict, 409},
		{errors.New("database unavailable: secret"), 500},
	}
	for _, test := range tests {
		status, message := reservationPolicyErrorResponse(test.err)
		if status != test.status {
			t.Fatalf("status = %d, want %d", status, test.status)
		}
		if status == 500 && message != "no se pudo publicar la politica de reservas" {
			t.Fatalf("internal detail leaked: %q", message)
		}
	}
}
