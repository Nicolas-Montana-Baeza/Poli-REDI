package repositories

import (
	"testing"

	"poli-redi-api/internal/models"
)

func TestPolicyRequiresGroupConfirmationForCancha2(t *testing.T) {
	policy := models.ReservationPolicy{GroupResourceIDs: []int{1, 2, 7}}

	if !policyRequiresGroupConfirmation(policy, 2) {
		t.Fatal("Cancha 2 (resource 2) must require group confirmation")
	}
	if policyRequiresGroupConfirmation(policy, 3) {
		t.Fatal("a reservable resource outside groupResourceIds must remain individual")
	}
	if policyRequiresGroupConfirmation(policy, 5) {
		t.Fatal("OPEN_USE resource 5 must remain outside group confirmation")
	}
}
