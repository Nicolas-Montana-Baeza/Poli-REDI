package services

import (
	"encoding/json"
	"strings"
	"testing"

	"poli-redi-api/internal/models"
)

func TestSanitizeMyReservationsPreservesOwnerAndRedactsParticipantAudience(t *testing.T) {
	input := []models.Reservation{
		{
			ID: 1, UserID: 7, UserFullName: "Usuario siete",
			UserEmail: "siete@example.test", UserRUT: "11111111-1",
		},
		{
			ID: 2, UserID: 99, UserFullName: "Propietario secreto",
			UserEmail: "owner@example.test", UserRUT: "22222222-2",
			CanEditTarget: true,
		},
	}

	got := sanitizeMyReservations(input, 7)
	if got[0].UserID != 7 || got[0].UserEmail != "siete@example.test" || got[0].UserRUT != "11111111-1" {
		t.Fatalf("owner identity was not preserved: %+v", got[0])
	}
	if got[1].UserID != 0 || got[1].UserFullName != "" || got[1].UserEmail != "" ||
		got[1].UserRUT != "" || got[1].CanEditTarget {
		t.Fatalf("participant audience received owner data: %+v", got[1])
	}

	payload, err := json.Marshal(got[1])
	if err != nil {
		t.Fatal(err)
	}
	body := string(payload)
	for _, forbidden := range []string{"userId", "userFullName", "userEmail", "userRut", "owner@example.test", "22222222-2"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("redacted response still serializes %q: %s", forbidden, body)
		}
	}
}
