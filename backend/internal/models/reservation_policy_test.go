package models

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestReservationPolicyPublicOmitsAdministrativeMetadata(t *testing.T) {
	payload, err := json.Marshal((ReservationPolicy{ID: 9, CreatedByUserID: pointer(2), ResourceIDs: []int{1}}).Public())
	if err != nil {
		t.Fatal(err)
	}
	text := string(payload)
	for _, forbidden := range []string{"\"id\"", "effectiveFrom", "createdByUserId", "createdAt"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("public payload leaks %s: %s", forbidden, text)
		}
	}
}

func pointer(value int) *int { return &value }
