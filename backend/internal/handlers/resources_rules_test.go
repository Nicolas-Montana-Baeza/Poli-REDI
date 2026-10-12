package handlers

import (
	"encoding/json"
	"testing"
)

func TestValidResourceReservationModes(t *testing.T) {
	valid := []string{
		"RESERVABLE",
		"OPEN_USE",
		"INFORMATIVE",
		"ADMIN_ONLY",
	}

	for _, mode := range valid {
		if !validResourceReservationMode(mode) {
			t.Fatalf("expected valid mode %s", mode)
		}
	}
}

func TestInvalidResourceReservationMode(t *testing.T) {
	if validResourceReservationMode("WHATEVER") {
		t.Fatal("unexpected valid reservation mode")
	}
}

func TestResourceCapacityAcceptsNull(t *testing.T) {
	var request updateResourceOperationalRequest

	err := json.Unmarshal(
		[]byte(`{"capacity":null}`),
		&request,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !request.Capacity.Present {
		t.Fatal("capacity should be present")
	}

	if request.Capacity.Value != nil {
		t.Fatal("capacity null should produce nil value")
	}
}

func TestResourceCapacityAcceptsPositiveInteger(t *testing.T) {
	var request updateResourceOperationalRequest

	err := json.Unmarshal(
		[]byte(`{"capacity":22}`),
		&request,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !request.Capacity.Present ||
		request.Capacity.Value == nil ||
		*request.Capacity.Value != 22 {

		t.Fatal("expected capacity 22")
	}
}
