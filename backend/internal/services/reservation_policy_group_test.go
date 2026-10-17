package services

import (
	"encoding/json"
	"testing"

	"poli-redi-api/internal/models"
)

func TestGroupResourcesAbsentVsEmpty(t *testing.T) {
	var absent, empty models.PublishReservationPolicyRequest

	if err := json.Unmarshal([]byte(`{}`), &absent); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(
		[]byte(`{"groupResources":[]}`), &empty,
	); err != nil {
		t.Fatal(err)
	}

	if absent.GroupResources != nil {
		t.Fatal("omitted groupResources must be nil")
	}
	if empty.GroupResources == nil ||
		len(*empty.GroupResources) != 0 {
		t.Fatal("explicit empty groupResources must be preserved")
	}

	hashAbsent, err := reservationPolicyPayloadHash(absent)
	if err != nil {
		t.Fatal(err)
	}
	hashEmpty, err := reservationPolicyPayloadHash(empty)
	if err != nil {
		t.Fatal(err)
	}

	if hashAbsent == hashEmpty {
		t.Fatal("omitted and empty must have different hashes")
	}
}

func TestNormalizePolicyGroupResources(t *testing.T) {
	rules := []models.PolicyGroupResource{
		{ResourceID: 3, MinimumParticipants: 12},
		{ResourceID: 1, MinimumParticipants: 10},
	}

	request := models.PublishReservationPolicyRequest{
		ResourceIDs:    []int{1, 3},
		GroupResources: &rules,
	}

	got, err := normalizePolicyGroupResources(request)
	if err != nil {
		t.Fatal(err)
	}

	if got.GroupResources == nil ||
		len(*got.GroupResources) != 2 ||
		(*got.GroupResources)[0].ResourceID != 1 ||
		(*got.GroupResources)[1].ResourceID != 3 {
		t.Fatalf("unexpected normalized resources: %+v", got.GroupResources)
	}
}

func TestPolicyGroupResourcesRejectInvalidRules(t *testing.T) {
	tests := []struct {
		name  string
		rules []models.PolicyGroupResource
	}{
		{
			"duplicate",
			[]models.PolicyGroupResource{
				{ResourceID: 1, MinimumParticipants: 10},
				{ResourceID: 1, MinimumParticipants: 12},
			},
		},
		{
			"outside policy",
			[]models.PolicyGroupResource{
				{ResourceID: 2, MinimumParticipants: 10},
			},
		},
		{
			"zero minimum",
			[]models.PolicyGroupResource{
				{ResourceID: 1, MinimumParticipants: 0},
			},
		},
		{
			"invalid id",
			[]models.PolicyGroupResource{
				{ResourceID: -1, MinimumParticipants: 10},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			request := models.PublishReservationPolicyRequest{
				ResourceIDs:    []int{1},
				GroupResources: &tc.rules,
			}

			_, err := normalizePolicyGroupResources(request)
			if err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestGroupResourcesCanonicalHash(t *testing.T) {
	firstRules := []models.PolicyGroupResource{
		{ResourceID: 3, MinimumParticipants: 12},
		{ResourceID: 1, MinimumParticipants: 10},
	}
	secondRules := []models.PolicyGroupResource{
		{ResourceID: 1, MinimumParticipants: 10},
		{ResourceID: 3, MinimumParticipants: 12},
	}

	first := models.PublishReservationPolicyRequest{
		ResourceIDs:    []int{1, 3},
		GroupResources: &firstRules,
	}
	second := models.PublishReservationPolicyRequest{
		ResourceIDs:    []int{1, 3},
		GroupResources: &secondRules,
	}

	var err error
	first, err = normalizePolicyGroupResources(first)
	if err != nil {
		t.Fatal(err)
	}
	second, err = normalizePolicyGroupResources(second)
	if err != nil {
		t.Fatal(err)
	}

	a, err := reservationPolicyPayloadHash(first)
	if err != nil {
		t.Fatal(err)
	}
	b, err := reservationPolicyPayloadHash(second)
	if err != nil {
		t.Fatal(err)
	}

	if a != b {
		t.Fatal("equivalent group rules must have same hash")
	}
}
