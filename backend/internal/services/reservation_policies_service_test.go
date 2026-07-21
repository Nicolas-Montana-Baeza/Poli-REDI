package services

import (
	"testing"

	"poli-redi-api/internal/models"
)

func TestUniquePositiveSortsAndDeduplicates(t *testing.T) {
	values := uniquePositive([]int{90, 30, 90, 60})
	want := []int{30, 60, 90}
	if len(values) != len(want) {
		t.Fatalf("uniquePositive() = %v", values)
	}
	for index := range want {
		if values[index] != want[index] {
			t.Fatalf("uniquePositive() = %v", values)
		}
	}
}

func TestReservationPolicyPayloadHashDetectsDivergentPayload(t *testing.T) {
	base := models.PublishReservationPolicyRequest{ReservableWindowDays: 7, AllowedDurations: []int{30, 60}, ResourceIDs: []int{1}}
	first, err := reservationPolicyPayloadHash(base)
	if err != nil {
		t.Fatal(err)
	}
	base.ReservableWindowDays = 8
	second, err := reservationPolicyPayloadHash(base)
	if err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Fatal("different payloads produced the same idempotency fingerprint")
	}
}
