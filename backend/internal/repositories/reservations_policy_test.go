package repositories

import (
	"testing"

	"poli-redi-api/internal/models"
)

func TestPolicyAllowsOnlyVersionedResources(t *testing.T) {
	policy := models.ReservationPolicy{ResourceIDs: []int{1, 3}}
	if !policyAllowsResource(policy, 3) {
		t.Fatal("versioned resource was rejected")
	}
	if policyAllowsResource(policy, 2) {
		t.Fatal("resource outside the policy snapshot was accepted")
	}
}

func TestInitialGroupReservationStatus(t *testing.T) {
	if got := initialGroupReservationStatus(10); got != models.ReservationStatusPending {
		t.Fatalf("minimum 10 = %s", got)
	}
	if got := initialGroupReservationStatus(1); got != models.ReservationStatusConfirmed {
		t.Fatalf("minimum 1 = %s", got)
	}
}
