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
