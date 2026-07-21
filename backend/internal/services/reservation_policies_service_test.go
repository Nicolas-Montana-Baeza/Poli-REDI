package services

import (
	"strings"
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

func TestPublishPolicyRejectsGroupResourceOutsideAllowedScope(t *testing.T) {
	request := models.PublishReservationPolicyRequest{ReservableWindowDays: 7, RequestFrequencyDays: 7, ConfirmationDeadlineMinutes: 60, MinimumParticipants: 10, OpeningMinute: 480, ClosingMinute: 1320, SlotIntervalMinutes: 15, AllowedDurations: []int{60}, ResourceIDs: []int{1}, GroupResourceIDs: []int{2}}
	_, _, err := PublishReservationPolicy(request, 1, "test-key")
	if err == nil || !strings.Contains(err.Error(), "deben pertenecer") {
		t.Fatalf("error = %v", err)
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
